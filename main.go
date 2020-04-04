package GrpcConnectionPool

import "time"

func main() {
	pool := InitConnectionPool()
	serviceList := InitPublishServiceList([]string{
		"go.service.Exam", "go.service.Marking", "go.service.Room", "go.service.StudentAndTeacher", "go.service.TestPaper", "go.service.Topic", "go.service.Utils",
	})
	serviceList.Subscript(pool)

	time.Sleep(100)

}
