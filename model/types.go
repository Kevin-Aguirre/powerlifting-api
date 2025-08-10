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
	Name string
	BirthDate string 
	Sex string 
	BirthYear string 
	Age string 
	Country string 
	State string 
	Equipment string 
	Division string 
	BodyweightKg string 
	WeightClassKg string 
	Squat1Kg string 
	Squat2Kg string 
	Squat3Kg string 
	Best3SquatKg string 
	Squat4Kg string
	Bench1Kg string
	Bench2Kg string
	Bench3Kg string
	Best3BenchKg string 
	Bench4Kg string
	Deadlift1Kg string
	Deadlift2Kg string
	Deadlift3Kg string
	Best3DeadliftKg string
	Deadlift4Kg string
	TotalKg string
	Event string
	Tested string
}

type Lifter struct {
	Name string
	PB map[string]*PersonalBest // string is division
	CompetitionResults []*LifterMeetResult
}

