//go:build ignore

#include "common.h"
#include "event_helpers.h"

SEC("tracepoint/syscalls/sys_enter_openat")
int tp_enter_openat(struct trace_event_raw_sys_enter *ctx)
{
    struct event_t *event = new_event();
    int flags = (int)ctx->args[2];

    if (!event)
        return 0;
    fill_event_identity(event, flags & O_CREAT ? EVENT_CREATE : EVENT_OPEN);
    bpf_probe_read_user_str(event->path, sizeof(event->path), (const char *)ctx->args[1]);
    return stash_event(event);
}

SEC("tracepoint/syscalls/sys_exit_openat")
int tp_exit_openat(struct trace_event_raw_sys_exit *ctx)
{
    return submit_event(ctx);
}
