package model

type PersonalBest struct {
	Squat float64
	Bench float64
	Deadlift float64
	Total float64
	Dots float64
}

type LifterMeetResult struct {
	Place string
	Federation string 
	Date string
	Location string 
	Competition string 
	Division string
	Age int
	Equip string 
	Class int
	Weight float64
	Squats []int // negative lift indicate no lift or failed attempst
	Benches []int // negative lift indicate no lift or failed attempst
	Deadlift []int // negative lift indicate no lift or failed attempst
	Total float64
	Dots float64
}

type Lifter struct {
	Name string
	PB map[string]*PersonalBest // string is division
	CompetitionResults []LifterMeetResult
}
