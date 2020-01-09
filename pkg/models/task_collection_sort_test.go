// Vikunja is a todo-list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"github.com/mohae/deepcopy"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"reflect"
	"testing"
)

func TestSortParamValidation(t *testing.T) {
	t.Run("Test valid order by", func(t *testing.T) {
		t.Run(orderAscending.String(), func(t *testing.T) {
			s := &sortParam{
				orderBy: orderAscending,
				sortBy:  "id",
			}
			err := s.validate()
			assert.NoError(t, err)
		})
		t.Run(orderDescending.String(), func(t *testing.T) {
			s := &sortParam{
				orderBy: orderDescending,
				sortBy:  "id",
			}
			err := s.validate()
			assert.NoError(t, err)
		})
	})
	t.Run("Test valid sort by", func(t *testing.T) {
		for _, test := range []sortProperty{
			taskPropertyID,
			taskPropertyText,
			taskPropertyDescription,
			taskPropertyDone,
			taskPropertyDoneAtUnix,
			taskPropertyDueDateUnix,
			taskPropertyCreatedByID,
			taskPropertyListID,
			taskPropertyRepeatAfter,
			taskPropertyPriority,
			taskPropertyStartDateUnix,
			taskPropertyEndDateUnix,
			taskPropertyHexColor,
			taskPropertyPercentDone,
			taskPropertyUID,
			taskPropertyCreated,
			taskPropertyUpdated,
		} {
			t.Run(test.String(), func(t *testing.T) {
				s := &sortParam{
					orderBy: orderAscending,
					sortBy:  test,
				}
				err := s.validate()
				assert.NoError(t, err)
			})
		}
	})
	t.Run("Test invalid order by", func(t *testing.T) {
		s := &sortParam{
			orderBy: "somethingInvalid",
			sortBy:  "id",
		}
		err := s.validate()
		assert.Error(t, err)
		assert.True(t, IsErrInvalidSortOrder(err))
	})
	t.Run("Test invalid sort by", func(t *testing.T) {
		s := &sortParam{
			orderBy: orderAscending,
			sortBy:  "somethingInvalid",
		}
		err := s.validate()
		assert.Error(t, err)
		assert.True(t, IsErrInvalidSortParam(err))
	})
}

var (
	task1 = &Task{
		ID:          1,
		Text:        "aaa",
		Description: "Lorem Ipsum",
		Done:        true,
		DoneAtUnix:  1543626000,
		ListID:      1,
		UID:         "JywtBPCESImlyKugvaZWrxmXAFAWXFISMeXYImEh",
		Created:     1543626724,
		Updated:     1543626724,
	}
	task2 = &Task{
		ID:            2,
		Text:          "bbb",
		Description:   "Arem Ipsum",
		Done:          true,
		DoneAtUnix:    1543626724,
		CreatedByID:   1,
		ListID:        2,
		PercentDone:   0.3,
		StartDateUnix: 1543626724,
		Created:       1553626724,
		Updated:       1553626724,
	}
	task3 = &Task{
		ID:          3,
		Text:        "ccc",
		DueDateUnix: 1583626724,
		Priority:    100,
		ListID:      3,
		HexColor:    "000000",
		PercentDone: 0.1,
		Updated:     1555555555,
	}
	task4 = &Task{
		ID:            4,
		Text:          "ddd",
		Priority:      1,
		StartDateUnix: 1643626724,
		ListID:        1,
	}
	task5 = &Task{
		ID:          5,
		Text:        "eef",
		Priority:    50,
		UID:         "shggzCHQWLhGNMNsOGOCOjcVkInOYjTAnORqTkdL",
		DueDateUnix: 1543636724,
		Updated:     1565555555,
	}
	task6 = &Task{
		ID:          6,
		Text:        "eef",
		DueDateUnix: 1543616724,
		RepeatAfter: 6400,
		CreatedByID: 2,
		HexColor:    "ffffff",
	}
	task7 = &Task{
		ID:            7,
		Text:          "mmmn",
		Description:   "Zoremis",
		StartDateUnix: 1544600000,
		EndDateUnix:   1584600000,
		UID:           "tyzCZuLMSKhwclJOsDyDcUdyVAPBDOPHNTBOLTcW",
	}
	task8 = &Task{
		ID:          8,
		Text:        "b123",
		EndDateUnix: 1544700000,
	}
	task9 = &Task{
		ID:            9,
		Done:          true,
		DoneAtUnix:    1573626724,
		Text:          "a123",
		RepeatAfter:   86000,
		StartDateUnix: 1544600000,
		EndDateUnix:   1544700000,
	}
	task10 = &Task{
		ID:          10,
		Text:        "zzz",
		Priority:    10,
		PercentDone: 1,
	}
)

type taskSortTestCase struct {
	name         string
	wantAsc      []*Task
	wantDesc     []*Task
	sortProperty sortProperty
}

var taskSortTestCases = []taskSortTestCase{
	{
		name:         "id",
		sortProperty: taskPropertyID,
		wantAsc: []*Task{
			task1,
			task2,
			task3,
			task4,
			task5,
			task6,
			task7,
			task8,
			task9,
			task10,
		},
		wantDesc: []*Task{
			task10,
			task9,
			task8,
			task7,
			task6,
			task5,
			task4,
			task3,
			task2,
			task1,
		},
	},
	{
		name:         "text",
		sortProperty: taskPropertyText,
		wantAsc: []*Task{
			task9,
			task1,
			task8,
			task2,
			task3,
			task4,
			task5,
			task6,
			task7,
			task10,
		},
		wantDesc: []*Task{
			task10,
			task7,
			task5,
			task6,
			task4,
			task3,
			task2,
			task8,
			task1,
			task9,
		},
	},
	{
		name:         "description",
		sortProperty: taskPropertyDescription,
		wantAsc: []*Task{
			task3,
			task4,
			task5,
			task6,
			task8,
			task9,
			task10,
			task2,
			task1,
			task7,
		},
		wantDesc: []*Task{
			task7,
			task1,
			task2,
			task3,
			task4,
			task5,
			task6,
			task8,
			task9,
			task10,
		},
	},
	{
		name:         "done",
		sortProperty: taskPropertyDone,
		wantAsc: []*Task{
			// These are not
			task3,
			task4,
			task5,
			task6,
			task7,
			task8,
			task10,
			// These are done
			task1,
			task2,
			task9,
		},
		wantDesc: []*Task{
			// These are done
			task1,
			task2,
			task9,
			// These are not
			task3,
			task4,
			task5,
			task6,
			task7,
			task8,
			task10,
		},
	},
	{
		name:         "done at",
		sortProperty: taskPropertyDoneAtUnix,
		wantAsc: []*Task{
			task3,
			task4,
			task5,
			task6,
			task7,
			task8,
			task10,
			task1,
			task2,
			task9,
		},
		wantDesc: []*Task{
			task9,
			task2,
			task1,
			task3,
			task4,
			task5,
			task6,
			task7,
			task8,
			task10,
		},
	},
	{
		name:         "due date",
		sortProperty: taskPropertyDueDateUnix,
		wantAsc: []*Task{
			task1,
			task2,
			task4,
			task7,
			task8,
			task9,
			task10,
			task6,
			task5,
			task3,
		},
		wantDesc: []*Task{
			task3,
			task5,
			task6,
			task1,
			task2,
			task4,
			task7,
			task8,
			task9,
			task10,
		},
	},
	{
		name:         "created by id",
		sortProperty: taskPropertyCreatedByID,
		wantAsc: []*Task{
			task1,
			task3,
			task4,
			task5,
			task7,
			task8,
			task9,
			task10,
			task2,
			task6,
		},
		wantDesc: []*Task{
			task6,
			task2,
			task1,
			task3,
			task4,
			task5,
			task7,
			task8,
			task9,
			task10,
		},
	},
	{
		name:         "list id",
		sortProperty: taskPropertyListID,
		wantAsc: []*Task{
			task5,
			task6,
			task7,
			task8,
			task9,
			task10,
			task1,
			task4,
			task2,
			task3,
		},
		wantDesc: []*Task{
			task3,
			task2,
			task1,
			task4,
			task5,
			task6,
			task7,
			task8,
			task9,
			task10,
		},
	},
	{
		name:         "repeat after",
		sortProperty: taskPropertyRepeatAfter,
		wantAsc: []*Task{
			task1,
			task2,
			task3,
			task4,
			task5,
			task7,
			task8,
			task10,
			task6,
			task9,
		},
		wantDesc: []*Task{
			task9,
			task6,
			task1,
			task2,
			task3,
			task4,
			task5,
			task7,
			task8,
			task10,
		},
	},
	{
		name:         "priority",
		sortProperty: taskPropertyPriority,
		wantAsc: []*Task{
			task1,
			task2,
			task6,
			task7,
			task8,
			task9,
			task4,
			task10,
			task5,
			task3,
		},
		wantDesc: []*Task{
			task3,
			task5,
			task10,
			task4,
			task1,
			task2,
			task6,
			task7,
			task8,
			task9,
		},
	},
	{
		name:         "start date",
		sortProperty: taskPropertyStartDateUnix,
		wantAsc: []*Task{
			task1,
			task3,
			task5,
			task6,
			task8,
			task10,
			task2,
			task7,
			task9,
			task4,
		},
		wantDesc: []*Task{
			task4,
			task7,
			task9,
			task2,
			task1,
			task3,
			task5,
			task6,
			task8,
			task10,
		},
	},
	{
		name:         "end date",
		sortProperty: taskPropertyEndDateUnix,
		wantAsc: []*Task{
			task1,
			task2,
			task3,
			task4,
			task5,
			task6,
			task10,
			task8,
			task9,
			task7,
		},
		wantDesc: []*Task{
			task7,
			task8,
			task9,
			task1,
			task2,
			task3,
			task4,
			task5,
			task6,
			task10,
		},
	},
	{
		name:         "hex color",
		sortProperty: taskPropertyHexColor,
		wantAsc: []*Task{
			task1,
			task2,
			task4,
			task5,
			task7,
			task8,
			task9,
			task10,
			task3,
			task6,
		},
		wantDesc: []*Task{
			task6,
			task3,
			task1,
			task2,
			task4,
			task5,
			task7,
			task8,
			task9,
			task10,
		},
	},
	{
		name:         "percent done",
		sortProperty: taskPropertyPercentDone,
		wantAsc: []*Task{
			task1,
			task4,
			task5,
			task6,
			task7,
			task8,
			task9,
			task3,
			task2,
			task10,
		},
		wantDesc: []*Task{
			task10,
			task2,
			task3,
			task1,
			task4,
			task5,
			task6,
			task7,
			task8,
			task9,
		},
	},
	{
		name:         "uid",
		sortProperty: taskPropertyUID,
		wantAsc: []*Task{
			task2,
			task3,
			task4,
			task6,
			task8,
			task9,
			task10,
			task1,
			task5,
			task7,
		},
		wantDesc: []*Task{
			task7,
			task5,
			task1,
			task2,
			task3,
			task4,
			task6,
			task8,
			task9,
			task10,
		},
	},
	{
		name:         "created",
		sortProperty: taskPropertyCreated,
		wantAsc: []*Task{
			task3,
			task4,
			task5,
			task6,
			task7,
			task8,
			task9,
			task10,
			task1,
			task2,
		},
		wantDesc: []*Task{
			task2,
			task1,
			task3,
			task4,
			task5,
			task6,
			task7,
			task8,
			task9,
			task10,
		},
	},
	{
		name:         "updated",
		sortProperty: taskPropertyUpdated,
		wantAsc: []*Task{
			task4,
			task6,
			task7,
			task8,
			task9,
			task10,
			task1,
			task2,
			task3,
			task5,
		},
		wantDesc: []*Task{
			task5,
			task3,
			task2,
			task1,
			task4,
			task6,
			task7,
			task8,
			task9,
			task10,
		},
	},
}

func TestTaskSort(t *testing.T) {

	assertTestSliceMatch := func(t *testing.T, got, want []*Task) {
		if !reflect.DeepEqual(got, want) {
			t.Error("Slices do not match in order")
			t.Error("Got\t| Want")
			for in, task := range got {
				fail := ""
				if task.ID != want[in].ID {
					fail = "wrong"
				}
				t.Errorf("\t%d\t| %d \t%s", task.ID, want[in].ID, fail)
			}
		}
	}

	for _, testCase := range taskSortTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Run("asc default", func(t *testing.T) {
				by := []*sortParam{
					{
						sortBy: testCase.sortProperty,
					},
				}

				got := deepcopy.Copy(testCase.wantAsc).([]*Task)

				// Destroy wanted order to obtain some slice we can sort
				rand.Shuffle(len(got), func(i, j int) {
					got[i], got[j] = got[j], got[i]
				})

				sortTasks(got, by)

				assertTestSliceMatch(t, got, testCase.wantAsc)
			})
			t.Run("asc", func(t *testing.T) {
				by := []*sortParam{
					{
						sortBy:  testCase.sortProperty,
						orderBy: orderAscending,
					},
				}

				got := deepcopy.Copy(testCase.wantAsc).([]*Task)

				// Destroy wanted order to obtain some slice we can sort
				rand.Shuffle(len(got), func(i, j int) {
					got[i], got[j] = got[j], got[i]
				})

				sortTasks(got, by)

				assertTestSliceMatch(t, got, testCase.wantAsc)
			})
			t.Run("desc", func(t *testing.T) {
				by := []*sortParam{
					{
						sortBy:  testCase.sortProperty,
						orderBy: orderDescending,
					},
				}

				got := deepcopy.Copy(testCase.wantDesc).([]*Task)

				// Destroy wanted order to obtain some slice we can sort
				rand.Shuffle(len(got), func(i, j int) {
					got[i], got[j] = got[j], got[i]
				})

				sortTasks(got, by)

				assertTestSliceMatch(t, got, testCase.wantDesc)
			})
		})
	}

	// Other cases
	t.Run("Order by Done Ascending and ID Descending", func(t *testing.T) {
		want := []*Task{
			// Not done
			task10,
			task8,
			task7,
			task6,
			task5,
			task4,
			task3,

			// Done
			task9,
			task2,
			task1,
		}
		sortParams := []*sortParam{
			{
				sortBy:  taskPropertyDone,
				orderBy: orderAscending,
			},
			{
				sortBy:  taskPropertyID,
				orderBy: orderDescending,
			},
		}

		got := deepcopy.Copy(want).([]*Task)

		// Destroy wanted order to obtain some slice we can sort
		rand.Shuffle(len(got), func(i, j int) {
			got[i], got[j] = got[j], got[i]
		})

		sortTasks(got, sortParams)

		assertTestSliceMatch(t, got, want)
	})
	t.Run("Order by Done Ascending and Text Descending", func(t *testing.T) {
		want := []*Task{
			// Not done
			task10,
			task7,
			task5,
			task6,
			task4,
			task3,
			task8,
			// Done
			task2,
			task1,
			task9,
		}
		sortParams := []*sortParam{
			{
				sortBy:  taskPropertyDone,
				orderBy: orderAscending,
			},
			{
				sortBy:  taskPropertyText,
				orderBy: orderDescending,
			},
		}

		got := deepcopy.Copy(want).([]*Task)

		// Destroy wanted order to obtain some slice we can sort
		rand.Shuffle(len(got), func(i, j int) {
			got[i], got[j] = got[j], got[i]
		})

		sortTasks(got, sortParams)

		assertTestSliceMatch(t, got, want)
	})
	t.Run("Order by Done Descending and Text Ascending", func(t *testing.T) {
		want := []*Task{
			// Done
			task9,
			task1,
			task2,
			// Not done
			task8,
			task3,
			task4,
			task5,
			task6,
			task7,
			task10,
		}
		sortParams := []*sortParam{
			{
				sortBy:  taskPropertyDone,
				orderBy: orderDescending,
			},
			{
				sortBy:  taskPropertyText,
				orderBy: orderAscending,
			},
		}

		got := deepcopy.Copy(want).([]*Task)

		// Destroy wanted order to obtain some slice we can sort
		rand.Shuffle(len(got), func(i, j int) {
			got[i], got[j] = got[j], got[i]
		})

		sortTasks(got, sortParams)

		assertTestSliceMatch(t, got, want)

	})
}
