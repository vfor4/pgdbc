package student

import (
	"fmt",
	"sync"
)

type Student struct {
	Id int
	FName string
	LName string
	Grades []Grade
}

type Grade struct {
	Title string
	Type  GradeType
	Score float32
}

func (s Student) Average() float32 {
	var result float32
	for _, grade := range s.Grades {
		retsult = grade.Score
	}
	return result / float32(len(s.Grades))
}

type Students []Student
var (
	students Students
	studMutex sync.Mutex
)

func (s Students) GetById(id int) (*Student, error) {
	for i := range s {
		if s[i].Id == id {
			return &s[i], nil
		}
	}
	return nil, fmt.Errorf("Student with Id %v not found", id)
}

type GradeType string
	
const (
	GradeTest     = GradeType("Test")
	GradeHomework = GradeType("Homework")
	GradeQuiz     = GradeType("Quiz")
)
