#!/bin/sh
set -e

#
# This script subsets our variable fonts,
# converts them to woff2 files and puts them in the
# fonts folder.
#
# We do have to update the font paths in the @font-face
# definitions manually since we use a checksum to make
#
# We use fonttools to create a partial instance of the
# variable font where we keep only our needed features.
# See more at:
# https://fonttools.readthedocs.io/en/latest/varLib/instancer.html
#
# fonttools requires python > 3.7. For up-to-date
# instructions see https://github.com/fonttools/fonttools#installation 
#
# Lot's of info was gathered from:
# https://markoskon.com/creating-font-subsets/ 
# https://barrd.dev/article/create-a-variable-font-subset-for-smaller-file-size/
#

ORIGINAL_FONTS="./originalMedia/fonts"
TEMP_FOLDER="./.subset-fonts-temp"
FONT_FOLDER="./src/assets/fonts"

err_report() {
	echo "Error on line $(caller)" >&2
}

trap err_report ERR

mkdir -p $TEMP_FOLDER

# the latin subset that google uses on GoogleFonts
# this is the same as the latin subset range that google uses on GoogleFonts
# see for examle the unicode-range definition here:
# https://fonts.googleapis.com/css2?family=Open+Sans
# Include the basic latin range plus latin-ext characters so glyphs with
# diacritics (used for example in Polish) are present in the generated fonts.
UNICODE_LATIN_SUBSET="U+0000-00FF,U+0100-017F,U+0131,U+0152-0153,U+02BB-02BC,\
U+02C6,U+02DA,U+02DC,U+2000-206F,U+2074,U+20AC,\
U+2122,U+2191,U+2193,U+2212,U+2215,U+FEFF,U+FFFD"

get_filename_without_type() {
	filename=$1
	dirname=$(dirname $filename)
	# Extract the file type using parameter expansion
	filetype=${filename##*.}
	basename=$(basename $filename .$filetype)
	echo $basename
}

# This function takes a font file and creates a subset of it using a specified set of unicode characters.
instance_and_subset () {
	# Define default arguments for the subsetter.
	DEFAULT_SUBSETTER_ARGS="--layout-features=* --unicodes=${UNICODE_LATIN_SUBSET}"

	# Assign function arguments to variables with more descriptive names.
	INPUT_FONT_FILE=$1
	INSTANCER_ARGS=$2
	OUTPUT_FONT_BASENAME=$3
	OUTPUT_FOLDER=$FONT_FOLDER

	# If the output font basename is not provided, use the input font file's basename as the output font basename.
	if [ -z "$OUTPUT_FONT_BASENAME" ]; then
		INPUT_FONT_BASENAME=$(get_filename_without_type $INPUT_FONT_FILE)
		OUTPUT_FONT_BASENAME=$INPUT_FONT_BASENAME
	fi

	# Use the default subsetter arguments if no custom arguments are provided.
	SUBSETTER_ARGS="${4:-$DEFAULT_SUBSETTER_ARGS}"

	CHECKSUM=$(
		# Concatenate the contents of the input font file, the instancer arguments, and the subsetter arguments
		printf "%s%s" "$(cat $INPUT_FONT_FILE)" "$INSTANCER_ARGS" "$SUBSETTER_ARGS" |
		# Calculate the Blake2b checksum of the concatenated string
		b2sum |
		# Extract the checksum from the output of b2sum (it's the first field)
		awk '{print $1}'
	)
	
	# Limit the checksum to 8 characters.
	CHECKSUM=$(echo "${CHECKSUM:0:8}")

	# Construct the output font's filename
	OUTPUT_FONT_BASENAME="${OUTPUT_FONT_BASENAME}_${CHECKSUM}"
	OUTPUT_FONT_FILE="${OUTPUT_FOLDER}/${OUTPUT_FONT_BASENAME}.woff2"

	# Check if the output font file already exists
	if test -f $OUTPUT_FONT_FILE; then
		echo "${OUTPUT_FONT_FILE} exists"
		return 0
	fi

	FONT_INSTANCE="${TEMP_FOLDER}/${OUTPUT_FONT_BASENAME}.ttf"

	if [ -n "$INSTANCER_ARGS" ]; then
		# If the INSTANCER_ARGS variable is set, use fonttools to create a font instance
		fonttools varLib.instancer --output $FONT_INSTANCE $INPUT_FONT_FILE $INSTANCER_ARGS
	else
		# Otherwise, just copy the input font file to the font instance file
		cp $INPUT_FONT_FILE $FONT_INSTANCE
	fi

	# Use pyftsubset to create a subset of the font instance and save it to the output font file
	pyftsubset $FONT_INSTANCE --output-file=$OUTPUT_FONT_FILE --flavor=woff2 $SUBSETTER_ARGS

	echo "${OUTPUT_FONT_BASENAME} subsetted."
}

echo ""
echo "###################################################"
echo "# Install required libs"
echo "###################################################"
echo ""

# Check if fonttools is available
if ! command -v fonttools >/dev/null 2>&1; then
	echo "fonttools not found, installing..."
	pip install fonttools
else
	echo "fonttools already available"
fi

# Check if pyftsubset is available (part of fonttools)
if ! command -v pyftsubset >/dev/null 2>&1; then
	echo "pyftsubset not found, installing fonttools..."
	pip install fonttools
else
	echo "pyftsubset already available"
fi

# Check if brotli is available
if ! python3 -c "import brotli" >/dev/null 2>&1; then
	echo "brotli not found, installing..."
	pip install brotli
else
	echo "brotli already available"
fi

echo ""
echo "###################################################"
echo "# Create a partial instance of the variable font"
echo "# where we keep only our needed features and then"
echo "# subset fonts with latin unicode range and export"
echo "# as woff2 file"
echo "###################################################"
echo ""

mkdir -p $TEMP_FOLDER

echo ""
echo "###################################################"
echo "# Collect existing font files for cleanup"
echo "###################################################"
echo ""

# Collect existing font files to remove later
OLD_FONT_FILES=$(find $FONT_FOLDER -name "*.woff2" -type f 2>/dev/null || true)
if [ -n "$OLD_FONT_FILES" ]; then
    echo "Found existing font files to remove after generation:"
    echo "$OLD_FONT_FILES"
else
    echo "No existing font files found"
fi

echo "\nOpen Sans"
# we drop the wdth axis for all

instance_and_subset "${ORIGINAL_FONTS}/OpenSans[wdth,wght].ttf" "wdth=drop wght=400:700" "OpenSans[wght]"

# we restrict the wght range
instance_and_subset "${ORIGINAL_FONTS}/OpenSans[wdth,wght].ttf" "wdth=drop wght=400" "OpenSans-Regular"
instance_and_subset "${ORIGINAL_FONTS}/OpenSans[wdth,wght].ttf" "wdth=drop wght=700" "OpenSans-Bold"

echo "\nOpen Sans Italic"
# we drop the wdth axis for all

instance_and_subset "${ORIGINAL_FONTS}/OpenSans-Italic[wdth,wght].ttf" "wdth=drop wght=400:700" "OpenSans-Italic[wght]"

# we restrict the wght range
instance_and_subset "${ORIGINAL_FONTS}/OpenSans-Italic[wdth,wght].ttf" "wdth=drop wght=400" "OpenSans-RegularItalic"
instance_and_subset "${ORIGINAL_FONTS}/OpenSans-Italic[wdth,wght].ttf" "wdth=drop wght=700" "OpenSans-BoldItalic"

echo "\nQuicksand"

instance_and_subset "${ORIGINAL_FONTS}/Quicksand[wght].ttf" "wght=400:700"

# we restrict the wght range
instance_and_subset "${ORIGINAL_FONTS}/Quicksand[wght].ttf" "wght=400" "Quicksand-Regular"
instance_and_subset "${ORIGINAL_FONTS}/Quicksand[wght].ttf" "wght=600" "Quicksand-SemiBold"
instance_and_subset "${ORIGINAL_FONTS}/Quicksand[wght].ttf" "wght=700" "Quicksand-Bold"

echo "\nSubsetting files complete"

echo ""
echo "###################################################"
echo "# Clean up old font files"
echo "###################################################"
echo ""

# Remove only the old font files we collected earlier
if [ -n "$OLD_FONT_FILES" ]; then
    echo "Removing old font files..."
    echo "$OLD_FONT_FILES" | while read -r file; do
        if [ -f "$file" ]; then
            echo "Removing: $file"
            rm -f "$file"
        fi
    done
else
    echo "No old font files to remove"
fi

echo ""
echo "###################################################"
echo "# Update fonts.scss with new font files"
echo "###################################################"
echo ""

FONTS_SCSS="./src/styles/fonts.scss"

echo "Updating $FONTS_SCSS with new font files..."

# Function to update font file references in SCSS
update_font_reference() {
    local pattern="$1"
    local new_file="$2"
    
    # Use sed to replace the font file reference, preserving the rest of the line
    sed -i "s|${pattern}_[a-f0-9]\{8\}\.woff2|${new_file}|g" "$FONTS_SCSS"
}

# Update each font file reference with the new checksum
for file in $FONT_FOLDER/*.woff2; do
    if [ -f "$file" ]; then
        basename=$(basename "$file")
        
        case $basename in
            OpenSans\[wght\]_*.woff2)
                update_font_reference "OpenSans\[wght\]" "$basename"
                ;;
            OpenSans-Italic\[wght\]_*.woff2)
                update_font_reference "OpenSans-Italic\[wght\]" "$basename"
                ;;
            Quicksand\[wght\]_*.woff2)
                update_font_reference "Quicksand\[wght\]" "$basename"
                ;;
            Quicksand-Regular_*.woff2)
                update_font_reference "Quicksand-Regular" "$basename"
                ;;
            Quicksand-SemiBold_*.woff2)
                update_font_reference "Quicksand-SemiBold" "$basename"
                ;;
            Quicksand-Bold_*.woff2)
                update_font_reference "Quicksand-Bold" "$basename"
                ;;
            OpenSans-Regular_*.woff2)
                update_font_reference "OpenSans-Regular" "$basename"
                ;;
            OpenSans-RegularItalic_*.woff2)
                update_font_reference "OpenSans-RegularItalic" "$basename"
                ;;
            OpenSans-Bold_*.woff2)
                update_font_reference "OpenSans-Bold" "$basename"
                ;;
            OpenSans-BoldItalic_*.woff2)
                update_font_reference "OpenSans-BoldItalic" "$basename"
                ;;
        esac
    fi
done

echo "fonts.scss updated with new font files"

# remove temp folder
rm -r $TEMP_FOLDER
