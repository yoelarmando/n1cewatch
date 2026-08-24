/* Minimal vmlinux.h stub for Windows dev — real BTF generated via bpftool on Linux.
   On Ubuntu build: bpftool btf dump file /sys/kernel/btf/vmlinux format c > bpf/vmlinux.h
*/
#pragma once
typedef unsigned int u32;
typedef unsigned long long u64;
struct trace_event_raw_sched_process_exec { const char *filename; };
struct trace_event_raw_sys_enter { unsigned long args[6]; };
struct trace_event_raw_inet_sock_set_state { int oldstate; int newstate; };
