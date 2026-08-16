#ifndef __EVENT_HELPERS_H
#define __EVENT_HELPERS_H

#include <bpf/bpf_core_read.h>

static __always_inline struct event_t *new_event(void)
{
    __u32 key = 0;
    struct event_t *event = bpf_map_lookup_elem(&scratch_event, &key);

    if (event)
        __builtin_memset(event, 0, sizeof(*event));
    return event;
}

static __always_inline void fill_event_identity(struct event_t *event, __u32 event_type)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u64 uid_gid = bpf_get_current_uid_gid();
    struct task_struct *task = (struct task_struct *)bpf_get_current_task_btf();

    event->timestamp_ns = bpf_ktime_get_ns();
    event->cgroup_id = bpf_get_current_cgroup_id();
    event->schema_version = EVENT_SCHEMA_VERSION;
    event->event_type = event_type;
    event->pid = (__u32)pid_tgid;
    event->tgid = pid_tgid >> 32;
    event->uid = (__u32)uid_gid;
    event->gid = uid_gid >> 32;
    event->ppid = BPF_CORE_READ(task, real_parent, tgid);
    event->mnt_ns_id = BPF_CORE_READ(task, nsproxy, mnt_ns, ns.inum);
    event->pid_ns_id = BPF_CORE_READ(task, nsproxy, pid_ns_for_children, ns.inum);
    bpf_get_current_comm(event->comm, sizeof(event->comm));
}

static __always_inline int stash_event(struct event_t *event)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    bpf_map_update_elem(&pending_events, &pid_tgid, event, BPF_ANY);
    return 0;
}

static __always_inline int submit_event(struct trace_event_raw_sys_exit *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    struct event_t *event = bpf_map_lookup_elem(&pending_events, &pid_tgid);

    if (!event)
        return 0;

    event->return_value = ctx->ret;
    bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, event, sizeof(*event));
    bpf_map_delete_elem(&pending_events, &pid_tgid);
    return 0;
}

#endif
