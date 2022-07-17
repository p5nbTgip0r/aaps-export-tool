package util

import (
	"aaps-export-tool/core"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"log"
	"strconv"
	"time"
)

// Objective is a groupings of tasks needed to be completed in order to unlock functionality in AAPS.
// Each objective has two preferences associated with it:
//
// When the objective was started: `Objectives_[NAME]_started = LONG_MILLIS` (default: 0)
//
// When the objective was completed: `Objectives_[NAME]_accomplished = LONG_MILLIS` (default: 0)
//
// An objective is considered completed when all of its tasks are completed.
//
// minimumDuration is actually considered to be a task in AAPS, but for the sake of simplicity it was split out into
// a field in this struct.
type Objective struct {
	Number          int
	Name            string
	minimumDuration time.Duration
	Tasks           []PreferenceTask
}

type PreferenceTask struct {
	key            string
	defaultValue   interface{}
	completedValue interface{}
}

// BooleanTask generates a boolean-based PreferenceTask, with the default value being `false` and the completed value being `true`
func BooleanTask(key string) *PreferenceTask {
	return &PreferenceTask{
		key:            key,
		defaultValue:   false,
		completedValue: true,
	}
}

// ExamTasks generates PreferenceTask's for exams. Exams are a type of task included in objectives.
//
// Each exam has two preferences associated with it:
//
// - whether the exam is completed: `ExamTask_[NAME] = BOOLEAN` (default: false)
//
// - whether the exam is locked out for an invalid answer: `DisabledTo_[NAME] = LONG_MILLIS` (default: 0)
func ExamTasks(names []string) []PreferenceTask {
	tasks := make([]PreferenceTask, 0, len(names)*2)
	for _, name := range names {
		tasks = append(tasks,
			[]PreferenceTask{
				{
					key:            "ExamTask_" + name,
					defaultValue:   false,
					completedValue: true,
				},
				{
					key:            "DisabledTo_" + name,
					defaultValue:   0,
					completedValue: 0,
				},
			}...,
		)
	}
	return tasks
}

// GetCompletedObjectives checks which objectives are already completed in the given preferences.
// This does not check task completion, only the objective's `accomplished` key.
func GetCompletedObjectives(contents []byte) []int {
	var completed []int

	for _, objective := range Objectives {
		started := gjson.GetBytes(contents, objective.StartedPrefKey()).String()
		accomplished := gjson.GetBytes(contents, objective.AccomplishedPrefKey()).String()

		startedInt, _ := strconv.Atoi(started)
		accomplishedInt, _ := strconv.Atoi(accomplished)

		startedTime := time.UnixMilli(int64(startedInt))
		accomplishedTime := time.UnixMilli(int64(accomplishedInt))

		isStarted := startedInt != 0
		isPastMinimumTime := objective.minimumDuration == 0 || isStarted && time.Now().Sub(startedTime) >= objective.minimumDuration
		isAccomplished := accomplishedInt != 0 && accomplishedTime.Before(time.Now())

		if isStarted && isPastMinimumTime && isAccomplished {
			completed = append(completed, objective.Number)
		}
	}

	return completed
}

func (obj *Objective) StartedPrefKey() string {
	return "Objectives_" + obj.Name + "_started"
}

func (obj *Objective) AccomplishedPrefKey() string {
	return "Objectives_" + obj.Name + "_accomplished"
}

func (obj *Objective) GetCompletionTime() time.Time {
	return time.Now().Add(obj.minimumDuration * -1)
}

func (obj *Objective) Complete(preferencesJson []byte) []byte {
	completedTime := obj.GetCompletionTime()
	completedTimeLong := strconv.FormatInt(completedTime.UnixMilli(), 10)
	// mark the objective as completed
	out, _ := sjson.SetBytes(preferencesJson, obj.StartedPrefKey(), completedTimeLong)
	out, _ = sjson.SetBytes(out, obj.AccomplishedPrefKey(), completedTimeLong)

	if core.Verbose {
		log.Printf("Set time for objective \"%s\" to \"%s\" (%s)", obj.Name, completedTimeLong, completedTime.Format(time.RFC1123Z))
	}

	// mark tasks as completed
	for _, task := range obj.Tasks {
		val := fmt.Sprintf("%v", task.completedValue)
		out, _ = sjson.SetBytes(out, task.key, val)
		if core.Verbose {
			log.Printf("Set task preference \"%s\": \"%s\"", task.key, val)
		}
	}

	return out
}

// ObjectiveNumbersToObjects converts a slice of objective numbers to the equivalent Objective structs
func ObjectiveNumbersToObjects(nums []int) []*Objective {
	objs := make([]*Objective, len(nums))

	for i, num := range nums {
		objs[i] = &Objectives[num-1]
	}

	return objs
}

var (
	// Objectives declares all objectives in AAPS, along with the necessary data needed to change an objective's completed status
	Objectives = []Objective{
		{
			Number: 1,
			Name:   "config",
			Tasks: []PreferenceTask{
				*BooleanTask("ObjectivesbgIsAvailableInNS"),
				*BooleanTask("virtualpump_uploadstatus"),
				*BooleanTask("ObjectivespumpStatusIsAvailableInNS"),
			},
		},
		{
			Number: 2,
			Name:   "usage",
			Tasks: []PreferenceTask{
				*BooleanTask("ObjectivesProfileSwitchUsed"),
				*BooleanTask("ObjectivesDisconnectUsed"),
				*BooleanTask("ObjectivesReconnectUsed"),
				*BooleanTask("ObjectivesTempTargetUsed"),
				*BooleanTask("ObjectivesActionsUsed"),
				*BooleanTask("ObjectivesLoopUsed"),
				*BooleanTask("ObjectivesScaleUsed"),
			},
		},
		{
			Number: 3,
			Name:   "exam",
			Tasks: ExamTasks([]string{
				"basaltest",
				"breadgrams",
				"dia",
				"exercise",
				"exercise2",
				"extendedcarbs",
				"hypott",
				"ic",
				"insulin",
				"iob",
				"isf",
				"noisycgm",
				"nsclient",
				"objectives",
				"objectives2",
				"otherMedicationWarning",
				"prerequisites",
				"prerequisites2",
				"profileswitch",
				"profileswitch2",
				"profileswitch4",
				"profileswitchtime",
				"pumpdisconnect",
				"sensitivity",
				"troubleshooting",
				"update",
				"wrongcarbs",
				"wronginsulin",
			}),
		},
		{
			Number: 4,
			Name:   "openloop",
			// 7 days
			minimumDuration: time.Hour * 24 * 7,
			Tasks: []PreferenceTask{
				{
					key:          "ObjectivesmanualEnacts",
					defaultValue: 0,
					// min amount needed per:
					// https://github.com/nightscout/AndroidAPS/blob/23207a275f97d1db1d993eae5122282920092602/app/src/main/java/info/nightscout/androidaps/plugins/constraints/objectives/objectives/Objective3.kt#L38
					completedValue: 20,
				},
			},
		},
		{
			Number: 5,
			Name:   "maxbasal",
		},
		{
			Number: 6,
			Name:   "maxiobzero",
			// 5 days
			minimumDuration: time.Hour * 24 * 5,
		},
		{
			Number: 7,
			Name:   "maxiob",
			// 1 day
			minimumDuration: time.Hour * 24,
		},
		{
			Number: 8,
			Name:   "autosens",
			// 7 days
			minimumDuration: time.Hour * 24 * 7,
		},

		// IMPORTANT: objective 9 (at the time it was AMA) was removed from AAPS, but the file names of objectives 9 (SMB)
		// and 10 (automations) were not changed. for the sake of clarity, i've tried to stick with the displayed
		// number of objectives in the GUI, but this discrepancy between the displayed number of the objectives and the
		// internal filename may be confusing if more are added. here's a map:
		//
		// filename (pref name)     | number in GUI | index (zero-indexed)
		// -------------------------|---------------|-----------------
		// Objective7.kt (autosens) | 8             | 7
		// Objective9.kt (smb)      | 9             | 8
		// Objective10.kt (auto)    | 10            | 9
		{
			Number: 9,
			Name:   "smb",
			// 28 days
			minimumDuration: time.Hour * 24 * 28,
		},
		{
			Number: 10,
			Name:   "auto",
			// 28 days
			minimumDuration: time.Hour * 24 * 28,
		},
	}
)
