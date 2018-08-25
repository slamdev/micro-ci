package internal

import (
	"github.com/slamdev/micro-ci/etc/schema"
	"k8s.io/api/core/v1"
	"testing"
)

func TestRunPipeline(t *testing.T) {
	err := RunPipeline(Pipeline{
		Jobs: []Job{
			{
				Name: "test",
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    "build",
							Image:   "alpine",
							Command: []string{"ls", "-la"},
						},
					},
				},
			},
		},
	}, schema.Commit{
		Author:   "slamdev",
		Branch:   "master",
		Revision: "b706fc1675e364f5cfd054b662560c7468f843b8",
	})
	if err != nil {
		t.Fatalf("Failed with error %+v", err)
	}
}
