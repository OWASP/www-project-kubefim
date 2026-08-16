//go:build ignore

#include "common.h"
#include "event_helpers.h"

static __always_inline int enter_chmod(struct trace_event_raw_sys_enter *ctx, const char *path)
{
    struct event_t *event = new_event();

    if (!event)
        return 0;
    fill_event_identity(event, EVENT_CHMOD);
    bpf_probe_read_user_str(event->path, sizeof(event->path), path);
    return stash_event(event);
}

SEC("tracepoint/syscalls/sys_enter_chmod")
int tp_enter_chmod(struct trace_event_raw_sys_enter *ctx)
{
    return enter_chmod(ctx, (const char *)ctx->args[0]);
}

SEC("tracepoint/syscalls/sys_exit_chmod")
int tp_exit_chmod(struct trace_event_raw_sys_exit *ctx)
{
    return submit_event(ctx);
}

SEC("tracepoint/syscalls/sys_enter_fchmodat")
int tp_enter_fchmodat(struct trace_event_raw_sys_enter *ctx)
{
    return enter_chmod(ctx, (const char *)ctx->args[1]);
}

SEC("tracepoint/syscalls/sys_exit_fchmodat")
int tp_exit_fchmodat(struct trace_event_raw_sys_exit *ctx)
{
    return submit_event(ctx);
}

SEC("tracepoint/syscalls/sys_enter_fchmodat2")
int tp_enter_fchmodat2(struct trace_event_raw_sys_enter *ctx)
{
    return enter_chmod(ctx, (const char *)ctx->args[1]);
}

SEC("tracepoint/syscalls/sys_exit_fchmodat2")
int tp_exit_fchmodat2(struct trace_event_raw_sys_exit *ctx)
{
    return submit_event(ctx);
}
