import type {VennMask} from './vennRegions'

export type Point = {x: number, y: number}
export type Circle = Point & {r: number}
// SVG text-anchor, so a title beside a circle grows away from it.
export type TitlePoint = Point & {align: 'start' | 'middle' | 'end'}

export type EulerLayout = {
	width: number,
	height: number,
	circles: Circle[],
	// Where each circle's title goes: just outside the circle, clear of the others where possible.
	titles: TitlePoint[],
	// One point per placed region, inside exactly the circles of its mask. Hiding circles keeps
	// every point correct for the circles that stay visible, so the layout never has to change.
	anchors: Map<VennMask, Point>,
	// Room around each anchor, in pixels, before it hits a circle edge.
	margins: Map<VennMask, number>,
	// Regions with tasks the layout could not give a spot of their own.
	unplaced: VennMask[],
	// Spot for the tasks without any visible label, in a strip below the diagram.
	outside: Point,
}

// Layout units: the biggest circle has this radius before the diagram is scaled to the canvas.
const UNIT = 100
// Smallest circle relative to the biggest, so a label with one task stays readable.
const MIN_RADIUS_RATIO = 0.32
const ITERATIONS = 400
const MAX_ROUNDS = 4
// Fewer starts for big clusters keeps the worst case fast; the best of them is usually found early.
const RANDOM_STARTS = 30
const BIG_CLUSTER = 6
const BIG_CLUSTER_STARTS = 10
const GRID_STEP = 4
// Room a region needs for its count bubble, in layout units.
const MIN_MARGIN = 14
// Space between circles that share no task.
const GAP = 18
// Space around a cluster when packing, with extra room on top for the circle titles.
const PACK_GAP = 28
const TITLE_ROOM = 30
const PADDING = 16
const OUTSIDE_STRIP = 56
// Keeps a lone label from filling the whole canvas.
const MAX_SCALE = 1.7
const MAX_HEIGHT = 560

// Lays out an Euler diagram over every label at once: one circle per label, its area following
// the number of tasks, overlapping only where tasks share those labels. Labels that share tasks
// form a cluster; clusters and lone labels are laid out on their own and packed side by side.
// The result is deterministic, and the view only shows or hides circles on top of it.
export function layoutEuler(setSizes: number[], occupiedMasks: VennMask[], width: number): EulerLayout {
	const n = setSizes.length
	const masks = [...new Set(occupiedMasks)].filter(mask => mask !== 0)
	const maxSize = Math.max(1, ...setSizes)
	const radii = setSizes.map(size => UNIT * Math.max(MIN_RADIUS_RATIO, Math.sqrt(size / maxSize)))

	const clusters = findClusters(n, masks).map(sets => layoutCluster(sets, radii, masks))
	const packed = pack(clusters)

	const scale = Math.min(
		MAX_SCALE,
		(width - 2 * PADDING) / Math.max(1, packed.width),
		(MAX_HEIGHT - 2 * PADDING - OUTSIDE_STRIP) / Math.max(1, packed.height),
	)
	const height = Math.max(240, Math.ceil(packed.height * scale + 2 * PADDING + OUTSIDE_STRIP))
	const offsetX = (width - packed.width * scale) / 2
	const offsetY = PADDING + (height - OUTSIDE_STRIP - 2 * PADDING - packed.height * scale) / 2
	const toCanvas = (p: Point): Point => ({x: offsetX + p.x * scale, y: offsetY + p.y * scale})

	const circles: Circle[] = Array(n)
	const anchors = new Map<VennMask, Point>()
	const margins = new Map<VennMask, number>()
	const unplaced: VennMask[] = []
	for (const {cluster, at} of packed.placements) {
		const shift = (p: Point) => toCanvas({x: p.x + at.x, y: p.y + at.y})
		cluster.sets.forEach((set, local) => {
			circles[set] = {...shift(cluster.circles[local]), r: cluster.circles[local].r * scale}
		})
		for (const [mask, point] of cluster.anchors) {
			anchors.set(mask, shift(point))
			margins.set(mask, cluster.margins.get(mask)! * scale)
		}
		unplaced.push(...cluster.unplaced)
	}

	return {
		width,
		height,
		circles,
		titles: placeTitles(circles, {width, height: height - OUTSIDE_STRIP}),
		anchors,
		margins,
		unplaced,
		outside: {x: 95, y: height - 22},
	}
}

type Cluster = {
	sets: number[],
	circles: Circle[],
	anchors: Map<VennMask, Point>,
	margins: Map<VennMask, number>,
	unplaced: VennMask[],
	bounds: {minX: number, minY: number, maxX: number, maxY: number},
}

// Groups sets into clusters of sets connected by shared tasks (union-find).
function findClusters(n: number, masks: VennMask[]): number[][] {
	const parent = [...Array(n).keys()]
	const root = (i: number): number => parent[i] === i ? i : (parent[i] = root(parent[i]))
	for (const mask of masks) {
		const bits = [...Array(n).keys()].filter(i => mask & (1 << i))
		for (const bit of bits.slice(1)) {
			parent[root(bit)] = root(bits[0])
		}
	}
	const groups = new Map<number, number[]>()
	for (let i = 0; i < n; i++) {
		groups.set(root(i), [...(groups.get(root(i)) ?? []), i])
	}
	return [...groups.values()]
}

function layoutCluster(sets: number[], radii: number[], allMasks: VennMask[]): Cluster {
	// Work with local bit positions so the force layout only sees this cluster.
	const toLocal = (mask: VennMask) => sets.reduce((local, set, i) => mask & (1 << set) ? local | (1 << i) : local, 0)
	const toGlobal = (local: VennMask) => sets.reduce((mask, set, i) => local & (1 << i) ? mask | (1 << set) : mask, 0)
	const masks = allMasks.filter(mask => sets.some(set => mask & (1 << set))).map(toLocal)
	const localRadii = sets.map(set => radii[set])

	let best: Attempt
	if (sets.length === 1) {
		const circle = {x: 0, y: 0, r: localRadii[0]}
		best = {circles: [circle], anchors: new Map([[1, {x: 0, y: 0}]]), margins: new Map([[1, circle.r]]), unplaced: [], score: 0}
	} else {
		best = layoutConnected(localRadii, masks)
	}

	return {
		sets,
		circles: best.circles,
		anchors: new Map([...best.anchors].map(([local, point]) => [toGlobal(local), point])),
		margins: new Map([...best.margins].map(([local, margin]) => [toGlobal(local), margin])),
		unplaced: best.unplaced.map(toGlobal),
		bounds: {
			minX: Math.min(...best.circles.map(c => c.x - c.r)),
			minY: Math.min(...best.circles.map(c => c.y - c.r)),
			maxX: Math.max(...best.circles.map(c => c.x + c.r)),
			maxY: Math.max(...best.circles.map(c => c.y + c.r)),
		},
	}
}

type Attempt = {
	circles: Circle[],
	anchors: Map<VennMask, Point>,
	margins: Map<VennMask, number>,
	unplaced: VennMask[],
	score: number,
}

// The force layout can get stuck depending on where the circles start, so try a few
// deterministic starts and keep the best: first every region placed, then as few circles
// overlapping without a shared task as possible.
function layoutConnected(radii: number[], masks: VennMask[]): Attempt {
	const n = radii.length
	const share: boolean[][] = Array.from({length: n}, () => Array<boolean>(n).fill(false))
	for (const mask of masks) {
		for (let i = 0; i < n; i++) {
			for (let j = i + 1; j < n; j++) {
				if (mask & (1 << i) && mask & (1 << j)) {
					share[i][j] = share[j][i] = true
				}
			}
		}
	}

	let best: Attempt | null = null
	for (const start of startingPositions(n, share, 1.2 * UNIT, n > BIG_CLUSTER ? BIG_CLUSTER_STARTS : RANDOM_STARTS)) {
		const attempt = runAttempt(start.map((p, i) => ({...p, r: radii[i]})), masks, share)
		const score = attempt.unplaced.length * 1000 + falseOverlap(attempt.circles, share)
		if (best === null || score < best.score) {
			best = {...attempt, score}
		}
		if (score === 0) {
			break
		}
	}
	return best!
}

function startingPositions(n: number, share: boolean[][], ring: number, randomStarts: number): Point[][] {
	const onRing = (order: number[]) => {
		const positions: Point[] = Array(n)
		order.forEach((set, slot) => {
			const angle = (2 * Math.PI * slot) / n - Math.PI / 2
			positions[set] = {x: ring * Math.cos(angle), y: ring * Math.sin(angle)}
		})
		return positions
	}

	// Greedy ordering that puts sets sharing tasks next to each other on the ring.
	const order = [0]
	while (order.length < n) {
		const last = order[order.length - 1]
		const rest = [...Array(n).keys()].filter(i => !order.includes(i))
		order.push(rest.find(i => share[last][i]) ?? rest[0])
	}

	const random = mulberry32(n * 7919)
	const starts = [onRing([...Array(n).keys()]), onRing(order)]
	for (let k = 0; k < randomStarts; k++) {
		starts.push(Array.from({length: n}, () => ({
			x: (random() - 0.5) * 2 * ring,
			y: (random() - 0.5) * 2 * ring,
		})))
	}
	return starts
}

function runAttempt(circles: Circle[], masks: VennMask[], share: boolean[][]) {
	const n = circles.length
	// Extra pull between the sets of a region that did not get a spot in the previous round.
	const pull: number[][] = Array.from({length: n}, () => Array<number>(n).fill(0))
	let placed = findAnchors(circles, masks)

	for (let round = 0; round < MAX_ROUNDS; round++) {
		relax(circles, share, pull)
		placed = findAnchors(circles, masks)
		const missing = masks.filter(mask => !placed.anchors.has(mask))
		if (missing.length === 0) {
			break
		}
		for (const mask of missing) {
			for (let i = 0; i < n; i++) {
				for (let j = i + 1; j < n; j++) {
					if (mask & (1 << i) && mask & (1 << j)) {
						pull[i][j] = pull[j][i] = pull[i][j] + 0.15
					}
				}
			}
		}
	}

	return {circles, ...placed, unplaced: masks.filter(mask => !placed.anchors.has(mask))}
}

// How badly circles overlap that share no task, relative to their size.
function falseOverlap(circles: Circle[], share: boolean[][]) {
	let total = 0
	for (let i = 0; i < circles.length; i++) {
		for (let j = i + 1; j < circles.length; j++) {
			const depth = circles[i].r + circles[j].r - Math.hypot(circles[i].x - circles[j].x, circles[i].y - circles[j].y)
			if (!share[i][j] && depth > 0) {
				total += 10 + depth / Math.min(circles[i].r, circles[j].r)
			}
		}
	}
	return total
}

function mulberry32(seed: number) {
	let a = seed
	return () => {
		a = (a + 0x6D2B79F5) | 0
		let t = Math.imul(a ^ (a >>> 15), 1 | a)
		t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t
		return ((t ^ (t >>> 14)) >>> 0) / 4294967296
	}
}

function relax(circles: Circle[], share: boolean[][], pull: number[][]) {
	const n = circles.length
	for (let step = 0; step < ITERATIONS; step++) {
		const rate = 0.2 * (1 - step / ITERATIONS) + 0.02
		const moves = circles.map(() => ({x: 0, y: 0}))
		for (let i = 0; i < n; i++) {
			for (let j = i + 1; j < n; j++) {
				const a = circles[i]
				const b = circles[j]
				const dx = b.x - a.x
				const dy = b.y - a.y
				const d = Math.hypot(dx, dy) || 0.01
				let force = 0
				if (share[i][j]) {
					// Overlap deeply enough that regions shared with a third set have room.
					const depth = Math.min(0.95, 0.6 + pull[i][j]) * Math.min(a.r, b.r)
					force = d - (a.r + b.r - depth)
				} else if (d < a.r + b.r + GAP) {
					// Push harder than shared sets pull, so unrelated circles do not end up touching.
					force = 1.5 * (d - (a.r + b.r + GAP))
				}
				const fx = (force * dx) / d / 2
				const fy = (force * dy) / d / 2
				moves[i].x += fx
				moves[i].y += fy
				moves[j].x -= fx
				moves[j].y -= fy
			}
		}
		circles.forEach((c, i) => {
			c.x += rate * (moves[i].x - 0.05 * c.x)
			c.y += rate * (moves[i].y - 0.05 * c.y)
		})
	}
}

function findAnchors(circles: Circle[], masks: VennMask[]) {
	const wanted = new Set(masks)
	const anchors = new Map<VennMask, Point>()
	const margins = new Map<VennMask, number>()
	const minX = Math.min(...circles.map(c => c.x - c.r))
	const maxX = Math.max(...circles.map(c => c.x + c.r))
	const minY = Math.min(...circles.map(c => c.y - c.r))
	const maxY = Math.max(...circles.map(c => c.y + c.r))
	for (let y = minY; y <= maxY; y += GRID_STEP) {
		for (let x = minX; x <= maxX; x += GRID_STEP) {
			let mask = 0
			let margin = Infinity
			circles.forEach((c, i) => {
				const d = Math.hypot(x - c.x, y - c.y)
				if (d < c.r) {
					mask |= 1 << i
				}
				margin = Math.min(margin, Math.abs(d - c.r))
			})
			if (wanted.has(mask) && margin >= MIN_MARGIN && margin > (margins.get(mask) ?? 0)) {
				anchors.set(mask, {x, y})
				margins.set(mask, margin)
			}
		}
	}
	return {anchors, margins}
}

// Shelf packing: biggest clusters first, left to right, wrapping into rows of a width that keeps
// the diagram roughly as wide as the canvas is tall.
function pack(clusters: Cluster[]) {
	const sizes = clusters.map(cluster => ({
		cluster,
		width: cluster.bounds.maxX - cluster.bounds.minX,
		height: cluster.bounds.maxY - cluster.bounds.minY + TITLE_ROOM,
	}))
	const totalArea = sizes.reduce((sum, s) => sum + (s.width + PACK_GAP) * (s.height + PACK_GAP), 0)
	const rowWidth = Math.max(...sizes.map(s => s.width), Math.sqrt(totalArea * 1.6))
	const order = sizes.map((s, i) => ({...s, i})).sort((a, b) => b.height - a.height || a.i - b.i)

	type Row = {items: typeof order, width: number, height: number}
	const rows: Row[] = []
	for (const item of order) {
		const row = rows[rows.length - 1]
		if (row && row.width + PACK_GAP + item.width <= rowWidth) {
			row.items.push(item)
			row.width += PACK_GAP + item.width
			row.height = Math.max(row.height, item.height)
		} else {
			rows.push({items: [item], width: item.width, height: item.height})
		}
	}

	const width = Math.max(...rows.map(row => row.width))
	const placements: {cluster: Cluster, at: Point}[] = []
	let y = 0
	for (const row of rows) {
		// Centre each row and each cluster vertically within its row.
		let x = (width - row.width) / 2
		for (const item of row.items) {
			const top = y + TITLE_ROOM + (row.height - item.height) / 2
			placements.push({
				cluster: item.cluster,
				at: {x: x - item.cluster.bounds.minX, y: top - item.cluster.bounds.minY},
			})
			x += item.width + PACK_GAP
		}
		y += row.height + PACK_GAP
	}
	return {width, height: y - PACK_GAP, placements}
}

const TITLE_ANGLES = [-90, -60, -120, -30, -150, 0, 180, 30, 150, 60, 120, 90].map(deg => (deg * Math.PI) / 180)

function placeTitles(circles: Circle[], box: {width: number, height: number}): TitlePoint[] {
	return circles.map((c, i) => {
		let best: TitlePoint | null = null
		let bestScore = -Infinity
		for (const angle of TITLE_ANGLES) {
			const cos = Math.cos(angle)
			const align = cos < -0.3 ? 'end' : cos > 0.3 ? 'start' : 'middle'
			const p = {x: c.x + (c.r + 8) * cos, y: c.y + (c.r + 12) * Math.sin(angle) + 4, align} as const
			if ((align === 'middle' && (p.x < 40 || p.x > box.width - 40)) ||
				(align === 'end' && p.x < 90) || (align === 'start' && p.x > box.width - 90) ||
				p.y < 12 || p.y > box.height) {
				continue
			}
			// Room to the nearest other circle; angles earlier in the list (on top) win ties.
			const room = Math.min(Infinity, ...circles.filter((_, j) => j !== i).map(o => Math.hypot(p.x - o.x, p.y - o.y) - o.r))
			const score = Math.min(room, 40)
			if (score > bestScore) {
				best = p
				bestScore = score
			}
		}
		return best ?? {x: c.x, y: Math.max(12, c.y - c.r - 8), align: 'middle'}
	})
}
