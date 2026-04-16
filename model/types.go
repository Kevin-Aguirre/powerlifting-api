package model

type Meet struct {
	Federation  string `json:"federation,omitempty"`
	Date        string `json:"date,omitempty"`
	MeetCountry string `json:"meetCountry,omitempty"`
	MeetState   string `json:"meetState,omitempty"`
	MeetTown    string `json:"meetTown,omitempty"`
	MeetName    string `json:"meetName,omitempty"`
	RuleSet     string `json:"ruleSet,omitempty"`
}

type PersonalBest struct {
	Squat    float64 `json:"squat,omitempty"`
	Bench    float64 `json:"bench,omitempty"`
	Deadlift float64 `json:"deadlift,omitempty"`
	Total    float64 `json:"total,omitempty"`
	Dots     float64 `json:"dots,omitempty"`
}

type LiftAttempts struct {
	Attempt1 float64 `json:"attempt1,omitempty"`
	Attempt2 float64 `json:"attempt2,omitempty"`
	Attempt3 float64 `json:"attempt3,omitempty"`
	Attempt4 float64 `json:"attempt4,omitempty"`
	Best     float64 `json:"best,omitempty"`
}

type LifterMeetResult struct {
	Place         string        `json:"place,omitempty"`
	Name          string        `json:"name"`
	BirthDate     string        `json:"birthDate,omitempty"`
	Sex           string        `json:"sex,omitempty"`
	BirthYear     int           `json:"birthYear,omitempty"`
	Age           float64       `json:"age,omitempty"`
	Country       string        `json:"country,omitempty"`
	State         string        `json:"state,omitempty"`
	Equipment     string        `json:"equipment,omitempty"`
	Division      string        `json:"division,omitempty"`
	BodyweightKg  float64       `json:"bodyweightKg,omitempty"`
	WeightClassKg string        `json:"weightClassKg,omitempty"`
	Squat         *LiftAttempts `json:"squat,omitempty"`
	Bench         *LiftAttempts `json:"bench,omitempty"`
	Deadlift      *LiftAttempts `json:"deadlift,omitempty"`
	TotalKg       float64       `json:"totalKg,omitempty"`
	Event         string        `json:"event,omitempty"`
	Tested        string        `json:"tested,omitempty"`
}

type Lifter struct {
	Name               string                  `json:"name"`
	PB                 map[string]*PersonalBest `json:"pb,omitempty"`
	CompetitionResults []*LifterMeetResult      `json:"competitionResults,omitempty"`
}

type RecordHolder struct {
	Lift   float64 `json:"lift"`
	Lifter string  `json:"lifter"`
}

type Record struct {
	WeightClassKg string        `json:"weightClassKg"`
	Sex           string        `json:"sex"`
	Equipment     string        `json:"equipment"`
	Squat         *RecordHolder `json:"squat,omitempty"`
	Bench         *RecordHolder `json:"bench,omitempty"`
	Deadlift      *RecordHolder `json:"deadlift,omitempty"`
	Total         *RecordHolder `json:"total,omitempty"`
}
