package grade

import (
	"fmt"
	"sync"
)

type Student struct {
	ID        int
	FirstName string
	LastName  string
	Grades    []Grade
}

func (s Student) Avarage() float32 {
	var result float32
	for _, grade := range s.Grades {
		result += grade.Score
	}
	return result / float32(len(s.Grades))
}

type Students []Student

func (s Students) GetById(id int) (*Student, error) {
	for i := range s {
		if s[i].ID == id {
			return &s[i], nil
		}
	}
	return nil, fmt.Errorf("Student with ID %v not found", id)
}

var (
	students      Students
	studentsMutex sync.Mutex
)

type GradeType string

const (
	GradeTest     = GradeType("Test")
	GradeHomework = GradeType("GradeHomework")
	GradeQuiz     = GradeType("GradeQuiz")
)

type Grade struct {
	Title string
	Type  GradeType
	Score float32
}
