package output

import (
	"encoding/json"
	"io"
	"strings"

	"kubefim/internal/event"
)

const EventAPIVersion = "events.kubefim.org/v1alpha1"

type JSON struct {
	encoder *json.Encoder
}

func NewJSON(writer io.Writer) *JSON {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	return &JSON{encoder: encoder}
}

func (j *JSON) Write(value event.Event) error {
	record := jsonEvent{
		APIVersion:    EventAPIVersion,
		SchemaVersion: value.SchemaVersion,
		TimestampNS:   value.TimestampNS,
		Operation:     strings.ToLower(value.Operation.String()),
		Success:       value.Successful(),
		ReturnValue:   value.ReturnValue,
		Process: jsonProcess{
			PID:  value.PID,
			TGID: value.TGID,
			PPID: value.PPID,
			UID:  value.UID,
			GID:  value.GID,
			Comm: value.Comm,
		},
		Target: jsonTarget{
			Path:            value.Path,
			DestinationPath: value.DestinationPath,
		},
		Linux: jsonLinux{
			CgroupID:       value.CgroupID,
			MountNamespace: value.MountNamespace,
			PIDNamespace:   value.PIDNamespace,
		},
	}
	if value.Container.ID != "" || value.Container.Host {
		record.Container = &jsonContainer{ID: value.Container.ID, Runtime: value.Container.Runtime, Host: value.Container.Host}
	}
	if value.Kubernetes.PodUID != "" {
		record.Kubernetes = &jsonKubernetes{
			Node: value.Kubernetes.Node, Namespace: value.Kubernetes.Namespace,
			PodName: value.Kubernetes.PodName, PodUID: value.Kubernetes.PodUID,
			ContainerName: value.Kubernetes.ContainerName, Image: value.Kubernetes.Image,
			ImageID: value.Kubernetes.ImageID,
		}
	}
	return j.encoder.Encode(record)
}

type jsonEvent struct {
	APIVersion    string          `json:"api_version"`
	SchemaVersion uint32          `json:"schema_version"`
	TimestampNS   uint64          `json:"timestamp_ns"`
	Operation     string          `json:"operation"`
	Success       bool            `json:"success"`
	ReturnValue   int64           `json:"return_value"`
	Process       jsonProcess     `json:"process"`
	Target        jsonTarget      `json:"target"`
	Linux         jsonLinux       `json:"linux"`
	Container     *jsonContainer  `json:"container,omitempty"`
	Kubernetes    *jsonKubernetes `json:"kubernetes,omitempty"`
}

type jsonContainer struct {
	ID      string `json:"id,omitempty"`
	Runtime string `json:"runtime,omitempty"`
	Host    bool   `json:"host"`
}

type jsonKubernetes struct {
	Node          string `json:"node"`
	Namespace     string `json:"namespace"`
	PodName       string `json:"pod_name"`
	PodUID        string `json:"pod_uid"`
	ContainerName string `json:"container_name"`
	Image         string `json:"image"`
	ImageID       string `json:"image_id"`
}

type jsonProcess struct {
	PID  uint32 `json:"pid"`
	TGID uint32 `json:"tgid"`
	PPID uint32 `json:"ppid"`
	UID  uint32 `json:"uid"`
	GID  uint32 `json:"gid"`
	Comm string `json:"comm"`
}

type jsonTarget struct {
	Path            string `json:"path"`
	DestinationPath string `json:"destination_path,omitempty"`
}

type jsonLinux struct {
	CgroupID       uint64 `json:"cgroup_id"`
	MountNamespace uint32 `json:"mount_namespace_id"`
	PIDNamespace   uint32 `json:"pid_namespace_id"`
}
