//go:build ignore

#include "common.h"
#include "event_helpers.h"

SEC("tracepoint/syscalls/sys_enter_unlinkat")
int tp_enter_unlinkat(struct trace_event_raw_sys_enter *ctx)
{
    struct event_t *event = new_event();

    if (!event)
        return 0;
    fill_event_identity(event, EVENT_DELETE);
    event->dirfd = (__s32)ctx->args[0];
    bpf_probe_read_user_str(event->path, sizeof(event->path), (const char *)ctx->args[1]);
    if (event->path[0] != '/')
        capture_path_base(event->dirfd, &event->base_inode, &event->base_device);
    return stash_event(event);
}

SEC("tracepoint/syscalls/sys_exit_unlinkat")
int tp_exit_unlinkat(struct trace_event_raw_sys_exit *ctx)
{
    return submit_event(ctx);
}
