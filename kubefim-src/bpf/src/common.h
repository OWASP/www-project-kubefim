#ifndef __COMMON_H
#define __COMMON_H

#define EVENT_SCHEMA_VERSION 1
#define TASK_COMM_LEN 16
#define EVENT_PATH_LEN 256

struct event_t {
    __u64 timestamp_ns;
    __u64 cgroup_id;
    __s64 return_value;
    __u32 schema_version;
    __u32 event_type;
    __u32 pid;
    __u32 tgid;
    __u32 ppid;
    __u32 uid;
    __u32 gid;
    __u32 mnt_ns_id;
    __u32 pid_ns_id;
    char comm[TASK_COMM_LEN];
    char path[EVENT_PATH_LEN];
    char destination_path[EVENT_PATH_LEN];
    __u32 reserved;
};

enum event_type_e {
    EVENT_OPEN = 1,
    EVENT_CREATE = 2,
    EVENT_DELETE = 3,
    EVENT_RENAME = 4,
    EVENT_CHMOD = 5,
};

#define O_CREAT 0100

#endif
