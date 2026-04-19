package model

import (
	"math"
	"net/http"
)

const kgToLbs = 2.20462

func roundTo2(f float64) float64 {
	return math.Round(f*100) / 100
}

func convertWeight(kg float64) float64 {
	if kg == 0 {
		return 0
	}
	return roundTo2(kg * kgToLbs)
}

// ConvertToLbs is the exported version of convertWeight for use in handlers.
func ConvertToLbs(kg float64) float64 {
	return convertWeight(kg)
}

// ParseUnit reads the "unit" query param. Returns true if lbs conversion is needed.
func ParseUnit(r *http.Request) bool {
	return r.URL.Query().Get("unit") == "lbs"
}

func (la *LiftAttempts) ToLbs() *LiftAttempts {
	if la == nil {
		return nil
	}
	return &LiftAttempts{
		Attempt1: convertWeight(la.Attempt1),
		Attempt2: convertWeight(la.Attempt2),
		Attempt3: convertWeight(la.Attempt3),
		Attempt4: convertWeight(la.Attempt4),
		Best:     convertWeight(la.Best),
	}
}

func (pb *PersonalBest) ToLbs() *PersonalBest {
	if pb == nil {
		return nil
	}
	return &PersonalBest{
		Squat:    convertWeight(pb.Squat),
		Bench:    convertWeight(pb.Bench),
		Deadlift: convertWeight(pb.Deadlift),
		Total:    convertWeight(pb.Total),
		Dots:     pb.Dots, // DOTS is unitless
	}
}

func (r *LifterMeetResult) ToLbs() *LifterMeetResult {
	if r == nil {
		return nil
	}
	return &LifterMeetResult{
		Place:          r.Place,
		Name:           r.Name,
		BirthDate:      r.BirthDate,
		Sex:            r.Sex,
		BirthYear:      r.BirthYear,
		Age:            r.Age,
		Country:        r.Country,
		State:          r.State,
		Equipment:      r.Equipment,
		Division:       r.Division,
		BodyweightKg:   convertWeight(r.BodyweightKg),
		WeightClassKg:  r.WeightClassKg,
		Squat:          r.Squat.ToLbs(),
		Bench:          r.Bench.ToLbs(),
		Deadlift:       r.Deadlift.ToLbs(),
		TotalKg:        convertWeight(r.TotalKg),
		Event:          r.Event,
		Tested:         r.Tested,
		MeetDate:       r.MeetDate,
		MeetFederation: r.MeetFederation,
		MeetName:       r.MeetName,
	}
}

func (l *Lifter) ToLbs() Lifter {
	converted := Lifter{
		Name:               l.Name,
		CompetitionResults: make([]*LifterMeetResult, len(l.CompetitionResults)),
	}

	if l.PB != nil {
		converted.PB = make(map[string]*PersonalBest, len(l.PB))
		for k, v := range l.PB {
			converted.PB[k] = v.ToLbs()
		}
	}

	for i, r := range l.CompetitionResults {
		converted.CompetitionResults[i] = r.ToLbs()
	}

	return converted
}
