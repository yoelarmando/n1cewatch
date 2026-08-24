// SPDX-License-Identifier: Dual MIT/GPL
// N1ceWatch CO-RE observer — tracepoints for all-Ubuntu 5.8+ with BTF.
// Falls back to auditd on 4.15 (Ubuntu 16.04/18.04). Inspired by Aurora.

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>

struct event {
    u32 pid;
    u32 ppid;
    u32 uid;
    u32 gid;
    u32 type; // 1=exec, 11=file, 3=net, 100=bpf
    char comm[16];
    char filename[256];
};

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 256 * 1024);
} events SEC(".maps");

// sched_process_exec tracepoint
SEC("tracepoint/sched/sched_process_exec")
int handle_exec(struct trace_event_raw_sched_process_exec *ctx) {
    struct event *e;
    e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
    if (!e) return 0;
    e->pid = bpf_get_current_pid_tgid() >> 32;
    e->uid = bpf_get_current_uid_gid() & 0xFFFFFFFF;
    e->gid = bpf_get_current_uid_gid() >> 32;
    e->type = 1;
    bpf_get_current_comm(&e->comm, sizeof(e->comm));
    bpf_probe_read_str(&e->filename, sizeof(e->filename), ctx->filename);
    bpf_ringbuf_submit(e, 0);
    return 0;
}

// openat tracepoint for file_event
SEC("tracepoint/syscalls/sys_enter_openat")
int handle_openat(struct trace_event_raw_sys_enter *ctx) {
    struct event *e;
    // Filter: only interesting paths in userspace, keep BPF fast
    e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
    if (!e) return 0;
    e->pid = bpf_get_current_pid_tgid() >> 32;
    e->type = 11;
    bpf_get_current_comm(&e->comm, sizeof(e->comm));
    // filename is arg2
    bpf_probe_read_user_str(&e->filename, sizeof(e->filename), (void*)ctx->args[1]);
    bpf_ringbuf_submit(e, 0);
    return 0;
}

// inet_sock_set_state for network
SEC("tracepoint/sock/inet_sock_set_state")
int handle_sock(struct trace_event_raw_inet_sock_set_state *ctx) {
    struct event *e;
    e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
    if (!e) return 0;
    e->pid = bpf_get_current_pid_tgid() >> 32;
    e->type = 3;
    bpf_get_current_comm(&e->comm, sizeof(e->comm));
    bpf_ringbuf_submit(e, 0);
    return 0;
}

// bpf syscall
SEC("tracepoint/syscalls/sys_enter_bpf")
int handle_bpf(struct trace_event_raw_sys_enter *ctx) {
    struct event *e;
    e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
    if (!e) return 0;
    e->pid = bpf_get_current_pid_tgid() >> 32;
    e->type = 100;
    bpf_get_current_comm(&e->comm, sizeof(e->comm));
    bpf_ringbuf_submit(e, 0);
    return 0;
}

char LICENSE[] SEC("license") = "Dual MIT/GPL";
