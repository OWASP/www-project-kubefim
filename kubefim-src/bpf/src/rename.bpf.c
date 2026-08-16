//go:build ignore

#include "common.h"
#include "event_helpers.h"

SEC("tracepoint/syscalls/sys_enter_renameat2")
int tp_enter_renameat2(struct trace_event_raw_sys_enter *ctx)
{
    struct event_t *event = new_event();

    if (!event)
        return 0;
    fill_event_identity(event, EVENT_RENAME);
    bpf_probe_read_user_str(event->path, sizeof(event->path), (const char *)ctx->args[1]);
    bpf_probe_read_user_str(event->destination_path, sizeof(event->destination_path),
                            (const char *)ctx->args[3]);
    return stash_event(event);
}

SEC("tracepoint/syscalls/sys_exit_renameat2")
int tp_exit_renameat2(struct trace_event_raw_sys_exit *ctx)
{
    return submit_event(ctx);
}
