package internal

import (
	"fmt"
	"github.com/slamdev/micro-ci/etc/schema"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	"strings"
)

func CreatePodTemplate(job Job, commit schema.Commit) core.Pod {
	pod := &core.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name:      job.Name + "-job-" + randomHash(5),
			Namespace: Namespace.Get(),
		},
		Spec: job.Spec,
	}
	pod.Spec.RestartPolicy = core.RestartPolicyNever
	pod.Spec.InitContainers = append([]core.Container{createCloneInitContainer(commit)}, pod.Spec.InitContainers...)
	repoVolume, repoVolumeMount := createRepoVolume()
	pod.Spec.Volumes = append(pod.Spec.Volumes, repoVolume)
	assignRepoVolumeMount(pod.Spec.Containers, repoVolumeMount)
	assignRepoVolumeMount(pod.Spec.InitContainers, repoVolumeMount)
	assignWorkingDir(pod.Spec.Containers)
	return *pod
}

func assignWorkingDir(containers []core.Container) {
	for idx := range containers {
		container := &containers[idx]
		if container.WorkingDir == "" {
			container.WorkingDir = "/opt/repo"
		}
	}
}

func assignRepoVolumeMount(containers []core.Container, volumeMount core.VolumeMount) {
	for idx := range containers {
		container := &containers[idx]
		container.VolumeMounts = append(container.VolumeMounts, volumeMount)
	}
}

func createCloneInitContainer(commit schema.Commit) core.Container {
	cloneCmd := fmt.Sprintf("git clone --branch %v %v /opt/repo", commit.Branch, RepoUrl.Get())
	cdCmd := "cd /opt/repo"
	checkoutCmd := "git reset --hard " + commit.Revision
	return core.Container{
		Name:  "clone",
		Image: "alpine/git",
		Command: []string{
			"sh", "-c",
			strings.Join([]string{cloneCmd, cdCmd, checkoutCmd}, " && "),
		},
	}
}

func createRepoVolume() (core.Volume, core.VolumeMount) {
	volume := core.Volume{
		Name: "repo",
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	}
	volumeMount := core.VolumeMount{
		Name:      "repo",
		MountPath: "/opt/repo",
	}
	return volume, volumeMount
}

func randomHash(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz1234567890"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
