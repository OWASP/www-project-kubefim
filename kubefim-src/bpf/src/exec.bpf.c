//go:build ignore

#include "common.h"
#include "event_helpers.h"

static __always_inline int enter_exec(const char *filename)
{
    struct event_t *event = new_event();

    if (!event)
        return 0;
    fill_event_identity(event, EVENT_EXEC);
    bpf_probe_read_user_str(event->path, sizeof(event->path), filename);
    return stash_event(event);
}

static __always_inline int exit_exec(struct trace_event_raw_sys_exit *ctx)
{
    /* sched_process_exec emits successful executions first and removes their
     * pending event. The syscall exit is also a fallback for kernels where the
     * scheduler tracepoint cannot be correlated; failed attempts only reach
     * this path. */
    return submit_event(ctx);
}

SEC("tracepoint/syscalls/sys_enter_execve")
int tp_enter_execve(struct trace_event_raw_sys_enter *ctx)
{
    return enter_exec((const char *)ctx->args[0]);
}

SEC("tracepoint/syscalls/sys_exit_execve")
int tp_exit_execve(struct trace_event_raw_sys_exit *ctx)
{
    return exit_exec(ctx);
}

SEC("tracepoint/syscalls/sys_enter_execveat")
int tp_enter_execveat(struct trace_event_raw_sys_enter *ctx)
{
    return enter_exec((const char *)ctx->args[1]);
}

SEC("tracepoint/syscalls/sys_exit_execveat")
int tp_exit_execveat(struct trace_event_raw_sys_exit *ctx)
{
    return exit_exec(ctx);
}

SEC("tracepoint/sched/sched_process_exec")
int tp_sched_process_exec(struct trace_event_raw_sched_process_exec *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u64 key = pid_tgid;
    struct event_t *event = bpf_map_lookup_elem(&pending_events, &key);

    /* A non-leader thread assumes the leader PID during exec. old_pid retains
     * the thread ID used to key the entry event. */
    if (!event && ctx->old_pid != (__u32)pid_tgid) {
        key = (pid_tgid & 0xffffffff00000000ULL) | (__u32)ctx->old_pid;
        event = bpf_map_lookup_elem(&pending_events, &key);
    }
    if (!event)
        return 0;

    event->return_value = 0;
    bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, event, sizeof(*event));
    bpf_map_delete_elem(&pending_events, &key);
    return 0;
}
