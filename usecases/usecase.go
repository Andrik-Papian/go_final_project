package usecases

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Andrik-Papian/go_final_project/model"
	"github.com/Andrik-Papian/go_final_project/repository"
)

const (
	year = 1

	sundayEU  = 0
	sundayRus = 7

	lastDayOfMonth     = -1
	predLastDayOfMonth = -2

	january  = 1
	december = 12

	maxDayOfMonth = 31

	timeFormatForSearch = "02.01.2006"
)

type TaskUsecase struct {
	db repository.TaskRepository
}

func NewTaskUsecase(db repository.TaskRepository) *TaskUsecase {
	return &TaskUsecase{db: db}
}

func (t *TaskUsecase) GetNextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("repeat is required")
	}

	dateTask, err := time.Parse(model.TimeFormat, date)
	if err != nil {
		return "", err
	}

	repeatString := strings.Split(repeat, " ")

	switch strings.ToLower(repeatString[0]) {
	case "d":
		if len(repeatString) < 2 {
			return "", fmt.Errorf("repeat should be at least two characters for days")
		}

		days, err := parseValue(repeatString[1])
		if err != nil {
			return "", err
		}
		dateTask = addDateTask(now, dateTask, 0, 0, days)
	case "y":
		dateTask = addDateTask(now, dateTask, year, 0, 0)
	case "w":
		if len(repeatString) < 2 {
			return "", fmt.Errorf("repeat should be at least more two characters for weeks")
		}

		dateTask, err = getDateTaskByWeek(now, dateTask, repeatString[1])
		if err != nil {
			return "", err
		}
	case "m":
		if len(repeatString) < 2 {
			return "", fmt.Errorf("repeat should be at least two characters for month")
		}

		dateTask, err = getDateTaskByMonth(now, dateTask, repeatString)
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("invalid character")
	}

	return dateTask.Format(model.TimeFormat), nil
}

func (t *TaskUsecase) CreateTask(task *model.Task, pastDay bool) (*model.TaskResp, error) {
	if pastDay {
		nextDate, err := t.GetNextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			return nil, err
		}

		task.Date = nextDate
	}

	taskId, err := t.db.CreateTask(task)
	if err != nil {
		return nil, err
	}

	taskResp := model.NewTaskResp(taskId)

	return taskResp, nil
}

func (t *TaskUsecase) GetTasks(searchString string) (model.TasksResp, error) {
	date, err := time.Parse(timeFormatForSearch, searchString)
	if err == nil {
		return t.db.GetTasksByDate(date)
	}

	if searchString != "" {
		return t.db.GetTasksBySearchString(searchString)
	}

	return t.db.GetTasks()
}

func (t *TaskUsecase) GetTaskById(id string) (*model.Task, error) {
	return t.db.GetTaskById(id)
}

func (t *TaskUsecase) UpdateTask(task *model.Task, pastDay bool) error {
	_, err := t.db.GetTaskById(task.Id)
	if err != nil {
		return err
	}

	if pastDay {
		nextDate, err := t.GetNextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			return err
		}

		task.Date = nextDate
	}

	return t.db.UpdateTask(task)
}

func (t *TaskUsecase) MakeTaskDone(id string) error {
	task, err := t.db.GetTaskById(id)
	if err != nil {
		return err
	}

	if task.Repeat == "" {
		return t.db.DeleteTask(id)
	}

	nextDate, err := t.GetNextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		return err
	}

	return t.db.MakeTaskDone(id, nextDate)
}

func (t *TaskUsecase) DeleteTask(id string) error {
	return t.db.DeleteTask(id)
}

func parseValue(num string) (int, error) {
	days, err := strconv.Atoi(num)
	if err != nil {
		return 0, err
	}

	if days >= 400 || days < 0 {
		return 0, fmt.Errorf("invalid value %d", days)
	}

	return days, nil
}

func addDateTask(now time.Time, dateTask time.Time, year int, month int, day int) time.Time {
	dateTask = dateTask.AddDate(year, month, day)

	for dateTask.Before(now) {
		dateTask = dateTask.AddDate(year, month, day)
	}

	return dateTask
}

func getDateTaskByWeek(now, dateTask time.Time, daysString string) (time.Time, error) {
	daysSlice := strings.Split(daysString, ",")
	daysOfWeek := regexp.MustCompile("[1-7]")

	daysOfWeekMap := make(map[int]bool)
	for _, day := range daysSlice {
		numberOfDay, err := strconv.Atoi(day)
		if err != nil {
			return dateTask, err
		}

		if len(daysOfWeek.FindAllString(day, -1)) == 0 {
			return dateTask, fmt.Errorf("invalid value %d day of the week", numberOfDay)
		}

		if numberOfDay == sundayRus {
			numberOfDay = sundayEU
		}
		daysOfWeekMap[numberOfDay] = true
	}

	for {
		if daysOfWeekMap[int(dateTask.Weekday())] {
			if now.Before(dateTask) {
				break
			}
		}
		dateTask = dateTask.AddDate(0, 0, 1)
	}

	return dateTask, nil
}

func getDateTaskByMonth(now, dateTask time.Time, repeat []string) (time.Time, error) {
	daysString := repeat[1]

	monthsString := ""
	if len(repeat) > 2 {
		monthsString = repeat[2]
	}

	daysSlice := strings.Split(daysString, ",")
	monthsSlice := strings.Split(monthsString, ",")

	daysMap := make(map[int]bool)
	for _, day := range daysSlice {
		numberOfDay, err := strconv.Atoi(day)
		if err != nil {
			return dateTask, err
		}
		if numberOfDay < predLastDayOfMonth || numberOfDay > maxDayOfMonth || numberOfDay == 0 {
			return dateTask, fmt.Errorf("invalid value %d day of the month", numberOfDay)
		}
		daysMap[numberOfDay] = true
	}

	monthsMap := make(map[int]bool)
	for _, month := range monthsSlice {
		if month == "" {
			continue
		}
		numberOfMonth, err := strconv.Atoi(month)
		if err != nil {
			return dateTask, err
		}
		if numberOfMonth < january || numberOfMonth > december {
			return dateTask, fmt.Errorf("invalid value %d month", numberOfMonth)
		}
		monthsMap[numberOfMonth] = true
	}

	for {
		if len(monthsMap) == 0 {
			break
		}

		if monthsMap[int(dateTask.Month())] {
			if now.Before(dateTask) {
				break
			}
		}
		dateTask = dateTask.AddDate(0, 0, 1)
	}

	for {
		lastDay := time.Date(dateTask.Year(), dateTask.Month()+1, 0, 0, 0, 0, 0, dateTask.Location()).Day()
		predLastDay := lastDay - 1

		key := dateTask.Day()
		switch {
		case lastDay == dateTask.Day():
			if _, ok := daysMap[lastDayOfMonth]; ok {
				key = lastDayOfMonth
			}
		case predLastDay == dateTask.Day():
			if _, ok := daysMap[predLastDayOfMonth]; ok {
				key = predLastDayOfMonth
			}
		}

		if daysMap[key] {
			if now.Before(dateTask) {
				break
			}
		}
		dateTask = dateTask.AddDate(0, 0, 1)
	}

	return dateTask, nil
}

/*package usecases

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Andrik-Papian/go_final_project/model"
	"github.com/Andrik-Papian/go_final_project/repository"
)

var _ Task = (*TaskUsecase)(nil)

const (
	year = 1

	sundayEU  = 0
	sundayRus = 7

	lastDayOfMonth     = -1
	predLastDayOfMonth = -2

	january  = 1
	december = 12

	maxDayOfMonth = 31

	timeFormatForSearch = "02.01.2006"
)

type TaskUsecase struct {
	db repository.TaskRepository
}

func NewTaskUsecase(db repository.TaskRepository) *TaskUsecase {
	return &TaskUsecase{db: db}
}

func (t *TaskUsecase) GetNextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("repeat is required")
	}

	dateTask, err := time.Parse(model.TimeFormat, date)
	if err != nil {
		return "", err
	}

	repeatString := strings.Split(repeat, " ")

	switch strings.ToLower(repeatString[0]) {
	case "d":
		if len(repeatString) < 2 {
			return "", fmt.Errorf("repeat should be at least two characters for days")
		}

		days, err := parseValue(repeatString[1])
		if err != nil {
			return "", err
		}
		dateTask = addDateTask(now, dateTask, 0, 0, days)
	case "y":
		dateTask = addDateTask(now, dateTask, year, 0, 0)
	case "w":
		if len(repeatString) < 2 {
			return "", fmt.Errorf("repeat should be at least more two characters for weeks")
		}

		dateTask, err = getDateTaskByWeek(now, dateTask, repeatString[1])
		if err != nil {
			return "", err
		}
	case "m":
		if len(repeatString) < 2 {
			return "", fmt.Errorf("repeat should be at least two characters for month")
		}

		dateTask, err = getDateTaskByMonth(now, dateTask, repeatString)
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("invalid character")
	}

	return dateTask.Format(model.TimeFormat), nil
}

func (t *TaskUsecase) CreateTask(task *model.Task, pastDay bool) (*model.TaskResp, error) {
	if pastDay {
		nextDate, err := t.GetNextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			return nil, err
		}

		task.Date = nextDate
	}

	taskId, err := t.db.CreateTask(task)
	if err != nil {
		return nil, err
	}

	taskResp := model.NewTaskResp(taskId)

	return taskResp, nil
}

func (t *TaskUsecase) GetTasks(searchString string) (model.TasksResp, error) {
	date, err := time.Parse(timeFormatForSearch, searchString)
	if err == nil {
		return t.db.GetTasksByDate(date)
	}

	if searchString != "" {
		return t.db.GetTasksBySearchString(searchString)
	}

	return t.db.GetTasks()
}

func (t *TaskUsecase) GetTaskById(id string) (*model.Task, error) {
	return t.db.GetTaskById(id)
}

func (t *TaskUsecase) UpdateTask(task *model.Task, pastDay bool) error {
	_, err := t.db.GetTaskById(task.Id)
	if err != nil {
		return err
	}

	if pastDay {
		nextDate, err := t.GetNextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			return err
		}

		task.Date = nextDate
	}

	return t.db.UpdateTask(task)
}

func (t *TaskUsecase) MakeTaskDone(id string) error {
	task, err := t.db.GetTaskById(id)
	if err != nil {
		return err
	}

	if task.Repeat == "" {
		return t.db.DeleteTask(id)
	}

	nextDate, err := t.GetNextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		return err
	}

	return t.db.MakeTaskDone(id, nextDate)
}

func (t *TaskUsecase) DeleteTask(id string) error {
	return t.db.DeleteTask(id)
}

func parseValue(num string) (int, error) {
	days, err := strconv.Atoi(num)
	if err != nil {
		return 0, err
	}

	if days >= 400 || days < 0 {
		return 0, fmt.Errorf("invalid value %d", days)
	}

	return days, nil
}

func addDateTask(now time.Time, dateTask time.Time, year int, month int, day int) time.Time {
	dateTask = dateTask.AddDate(year, month, day)

	for dateTask.Before(now) {
		dateTask = dateTask.AddDate(year, month, day)
	}

	return dateTask
}

func getDateTaskByWeek(now, dateTask time.Time, daysString string) (time.Time, error) {
	daysSlice := strings.Split(daysString, ",")
	daysOfWeek := regexp.MustCompile("[1-7]")

	daysOfWeekMap := make(map[int]bool)
	for _, day := range daysSlice {
		numberOfDay, err := strconv.Atoi(day)
		if err != nil {
			return dateTask, err
		}

		if len(daysOfWeek.FindAllString(day, -1)) == 0 {
			return dateTask, fmt.Errorf("invalid value %d day of the week", numberOfDay)
		}

		if numberOfDay == sundayRus {
			numberOfDay = sundayEU
		}
		daysOfWeekMap[numberOfDay] = true
	}

	for {
		if daysOfWeekMap[int(dateTask.Weekday())] {
			if now.Before(dateTask) {
				break
			}
		}
		dateTask = dateTask.AddDate(0, 0, 1)
	}

	return dateTask, nil
}

func getDateTaskByMonth(now, dateTask time.Time, repeat []string) (time.Time, error) {
	daysString := repeat[1]

	monthsString := ""
	if len(repeat) > 2 {
		monthsString = repeat[2]
	}

	daysSlice := strings.Split(daysString, ",")
	monthsSlice := strings.Split(monthsString, ",")

	daysMap := make(map[int]bool)
	for _, day := range daysSlice {
		numberOfDay, err := strconv.Atoi(day)
		if err != nil {
			return dateTask, err
		}
		if numberOfDay < predLastDayOfMonth || numberOfDay > maxDayOfMonth || numberOfDay == 0 {
			return dateTask, fmt.Errorf("invalid value %d day of the month", numberOfDay)
		}
		daysMap[numberOfDay] = true
	}

	monthsMap := make(map[int]bool)
	for _, month := range monthsSlice {
		if month == "" {
			continue
		}
		numberOfMonth, err := strconv.Atoi(month)
		if err != nil {
			return dateTask, err
		}
		if numberOfMonth < january || numberOfMonth > december {
			return dateTask, fmt.Errorf("invalid value %d month", numberOfMonth)
		}
		monthsMap[numberOfMonth] = true
	}

	for {
		if len(monthsMap) == 0 {
			break
		}

		if monthsMap[int(dateTask.Month())] {
			if now.Before(dateTask) {
				break
			}
		}
		dateTask = dateTask.AddDate(0, 0, 1)
	}

	for {
		lastDay := time.Date(dateTask.Year(), dateTask.Month()+1, 0, 0, 0, 0, 0, dateTask.Location()).Day()
		predLastDay := lastDay - 1

		key := dateTask.Day()
		switch {
		case lastDay == dateTask.Day():
			if _, ok := daysMap[lastDayOfMonth]; ok {
				key = lastDayOfMonth
			}
		case predLastDay == dateTask.Day():
			if _, ok := daysMap[predLastDayOfMonth]; ok {
				key = predLastDayOfMonth
			}
		}

		if daysMap[key] {
			if now.Before(dateTask) {
				break
			}
		}
		dateTask = dateTask.AddDate(0, 0, 1)
	}

	return dateTask, nil
}
*/
/*package usecases

import (
	"time"

	"github.com/Andrik-Papian/go_final_project/model"
	"github.com/Andrik-Papian/go_final_project/repository"
)

type TaskUseCase struct {
	TaskRepo repository.TaskRepo
}

func NewTaskUseCase(taskRepo repository.TaskRepo) *TaskUseCase {
	return &TaskUseCase{TaskRepo: taskRepo}
}

func (u *TaskUseCase) CreateTask(task *model.Task) error {
	_, err := u.TaskRepo.CreateTask(task)
	if err != nil {
		return err
	}
	return nil
}

func (u *TaskUseCase) GetTasks() (model.TasksResp, error) {
	return u.TaskRepo.GetTasks()
}

func (u *TaskUseCase) GetTasksBySearchString(searchString string) (model.TasksResp, error) {
	return u.TaskRepo.GetTasksBySearchString(searchString)
}

func (u *TaskUseCase) GetTasksByDate(searchDate time.Time) (model.TasksResp, error) {
	return u.TaskRepo.GetTasksByDate(searchDate)
}

func (u *TaskUseCase) GetTaskById(id string) (*model.Task, error) {
	return u.TaskRepo.GetTaskById(id)
}

func (u *TaskUseCase) UpdateTask(task *model.Task) error {
	return u.TaskRepo.UpdateTask(task)
}

func (u *TaskUseCase) MakeTaskDone(id string, date string) error {
	return u.TaskRepo.MakeTaskDone(id, date)
}

func (u *TaskUseCase) DeleteTask(id string) error {
	return u.TaskRepo.DeleteTask(id)
}
*/
