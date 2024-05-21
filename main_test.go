package main

import (
	"fmt"
	"slices"
	"testing"
)


func TestDoBattles(t *testing.T) {
	type testCase struct {
		users                  []user
		mv                     move
		expectedFightLocations []piece
	}
	tests := []testCase{
		{
			users: []user{
				{
					name: "Toussaint",
					pieces: []piece{
						{
							location: "San Domingo",
							name:     "Cavalry",
						},
						{
							location: "San Domingo",
							name:     "Infantry",
						},
					},
				},
				{
					name: "Napoleon",
					pieces: []piece{
						{
							location: "France",
							name:     "Infantry",
						},
						{
							location: "Russia",
							name:     "Infantry",
						},
					},
				},
				{
					name: "Washington",
					pieces: []piece{
						{
							location: "United States",
							name:     "Artillery",
						},
					},
				},
			},
			mv: move{
				userName: "Toussaint",
				piece: piece{
					location: "United States",
					name:     "Cavalry",
				},
			},
			expectedFightLocations: []piece{
				{
					location: "United States",
					name:     "Artillery",
				},
			},
		},
	}
	if true {
		tests = append(tests, []testCase{
			{
				users: []user{
					{
						name: "Toussaint",
						pieces: []piece{
							{
								location: "France",
								name:     "Cavalry",
							},
							{
								location: "San Domingo",
								name:     "Infantry",
							},
						},
					},
					{
						name: "Napoleon",
						pieces: []piece{
							{
								location: "France",
								name:     "Infantry",
							},
							{
								location: "Russia",
								name:     "Infantry",
							},
							{
								location: "United States",
								name:     "Cavalry",
							},
						},
					},
					{
						name: "Washington",
						pieces: []piece{
							{
								location: "United States",
								name:     "Artillery",
							},
						},
					},
					{
						name: "King George",
						pieces: []piece{
							{
								location: "United States",
								name:     "Infantry",
							},
							{
								location: "Great Britain",
								name:     "Infantry",
							},
						},
					},
				},
				mv: move{
					userName: "King George",
					piece: piece{
						location: "France",
						name:     "Infantry",
					},
				},
				expectedFightLocations: []piece{
					{
						location: "France",
						name:     "Cavalry",
					},
					{
						location: "France",
						name:     "Infantry",
					},
				},
			},
		}...)
	}

	for _, test := range tests {
		bufferedCh := make(chan move, 100)
		mover := user{}
		for _, u := range test.users {
			if u.name == test.mv.userName {
				mover = u
			}
		}
		if mover.name == "" {
			t.Errorf("Test Failed: user with name %v not found", test.mv.userName)
		}
		mover.march(test.mv.piece, bufferedCh)
		close(bufferedCh)

		subChans := []chan move{}
		for range test.users {
			subChans = append(subChans, make(chan move, 100))
		}
		distributeBattles(bufferedCh, subChans)
		for _, subChan := range subChans {
			close(subChan)
		}
		battles := []piece{}
		for i, u := range test.users {
			subChan := subChans[i]
			userBattles := u.doBattles(subChan)
			battles = append(battles, userBattles...)
		}

		if !slices.Equal(battles, test.expectedFightLocations) {
			t.Errorf(`
Test Failed:
  users:
%v
  move: %v
  =>
  expected battle pieces:
%v
  actual battle pieces:
%v
			`,
				formatSlice(test.users),
				test.mv,
				formatSlice(test.expectedFightLocations),
				formatSlice(battles),
			)
		} else {
			fmt.Printf(`
Test Passed:
  users:
%v
  move: %v
  =>
  expected battle pieces:
%v
  actual battle pieces:
%v
			`,
				formatSlice(test.users),
				test.mv,
				formatSlice(test.expectedFightLocations),
				formatSlice(battles),
			)
		}
	}
}

func TestMarch(t *testing.T) {
	type testCase struct {
		users                  []user
		mv                     move
		expectedFightLocations []piece
	}
	tests := []testCase{
		{
			users: []user{
				{
					name: "Toussaint",
					pieces: []piece{
						{
							location: "San Domingo",
							name:     "Cavalry",
						},
						{
							location: "San Domingo",
							name:     "Infantry",
						},
					},
				},
				{
					name: "Napoleon",
					pieces: []piece{
						{
							location: "France",
							name:     "Infantry",
						},
						{
							location: "Russia",
							name:     "Infantry",
						},
					},
				},
				{
					name: "Washington",
					pieces: []piece{
						{
							location: "United States",
							name:     "Artillery",
						},
					},
				},
			},
			mv: move{
				userName: "Toussaint",
				piece: piece{
					location: "United States",
					name:     "Cavalry",
				},
			},
			expectedFightLocations: []piece{
				{
					location: "United States",
					name:     "Artillery",
				},
			},
		},
	}
	if true {
		tests = append(tests, []testCase{
			{
				users: []user{
					{
						name: "Toussaint",
						pieces: []piece{
							{
								location: "San Domingo",
								name:     "Cavalry",
							},
							{
								location: "San Domingo",
								name:     "Infantry",
							},
						},
					},
					{
						name: "Napoleon",
						pieces: []piece{
							{
								location: "France",
								name:     "Infantry",
							},
							{
								location: "Russia",
								name:     "Infantry",
							},
							{
								location: "United States",
								name:     "Cavalry",
							},
						},
					},
					{
						name: "Washington",
						pieces: []piece{
							{
								location: "United States",
								name:     "Artillery",
							},
						},
					},
				},
				mv: move{
					userName: "Toussaint",
					piece: piece{
						location: "United States",
						name:     "Cavalry",
					},
				},
				expectedFightLocations: []piece{
					{
						location: "United States",
						name:     "Cavalry",
					},
					{
						location: "United States",
						name:     "Artillery",
					},
				},
			},
		}...)
	}

	for _, test := range tests {
		bufferedCh := make(chan move, 100)
		mover := user{}
		for _, u := range test.users {
			if u.name == test.mv.userName {
				mover = u
			}
		}
		if mover.name == "" {
			t.Errorf("Test Failed: user with name %v not found", test.mv.userName)
		}
		mover.march(test.mv.piece, bufferedCh)
		close(bufferedCh)
		output := doBattles(bufferedCh, test.users)
		if !slices.Equal(output, test.expectedFightLocations) {
			t.Errorf(`
Test Failed:
  users:
%v
  move: %v
  =>
  expected battle pieces:
%v
  actual battle pieces:
%v
			`,
				formatSlice(test.users),
				test.mv,
				formatSlice(test.expectedFightLocations),
				formatSlice(output),
			)
		} else {
			fmt.Printf(`
Test Passed:
  users:
%v
  move: %v
  =>
  expected battle pieces:
%v
  actual battle pieces:
%v
			`,
				formatSlice(test.users),
				test.mv,
				formatSlice(test.expectedFightLocations),
				formatSlice(output),
			)
		}
	}
}

func formatSlice[T any](slice []T) string {
	if slice == nil {
		return "nil"
	}
	output := ""
	for i, v := range slice {
		output += fmt.Sprintf("* %v", v)
		if i < len(slice)-1 {
			output += "\n"
		}
	}
	return output
}
