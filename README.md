Reproduce soft lock which is associated with "zap_pid_ns_processes" and zombie defunct processes

Example dmesg:
```
watchdog: BUG: soft lockup - CPU#0 stuck for 212s! [npm start:306207]
Modules linked in: veth nf_conntrack_netlink xt_conntrack nft_chain_nat xt_MASQUERADE nf_nat nf_conntrack nf_defrag_ipv6 nf_defrag_ipv4 xfrm_user xfrm_algo nft_counter xt_addrtype nft_compat nf_tables nfnetlink binfmt_misc nls_iso8859_1 intel_rapl_msr serio_raw intel_rapl_common hyperv_fb hv_balloon joydev mac_hid sch_fq_codel dm_multipath scsi_dh_rdac scsi_dh_emc scsi_dh_alua overlay iptable_filter ip6table_filter ip6_tables br_netfilter bridge stp llc arp_tables msr efi_pstore ip_tables x_tables autofs4 btrfs blake2b_generic zstd_compress raid10 raid456 async_raid6_recov async_memcpy async_pq async_xor async_tx xor raid6_pq libcrc32c raid1 raid0 multipath linear hyperv_drm drm_kms_helper syscopyarea sysfillrect sysimgblt fb_sys_fops crct10dif_pclmul cec hv_storvsc crc32_pclmul hid_generic hv_netvsc ghash_clmulni_intel scsi_transport_fc rc_core sha256_ssse3 hid_hyperv drm sha1_ssse3 hv_utils hid hyperv_keyboard aesni_intel crypto_simd cryptd hv_vmbus
CPU: 0 PID: 306207 Comm: npm start Tainted: G             L    5.15.0-107-generic #117-Ubuntu
Hardware name: Microsoft Corporation Virtual Machine/Virtual Machine, BIOS Hyper-V UEFI Release v4.1 04/06/2022
RIP: 0010:_raw_spin_unlock_irqrestore+0x25/0x30
Code: eb 8d cc cc cc 0f 1f 44 00 00 55 48 89 e5 e8 3a b8 36 ff 66 90 f7 c6 00 02 00 00 75 06 5d e9 e2 cb 22 00 fb 66 0f 1f 44 00 00 <5d> e9 d5 cb 22 00 0f 1f 44 00 00 0f 1f 44 00 00 55 48 89 e5 8b 07
RSP: 0018:ffffb15fc915bc60 EFLAGS: 00000206
RAX: 0000000000000001 RBX: ffffb15fc915bcf8 RCX: 0000000000000000
RDX: ffff9d4713f9c828 RSI: 0000000000000246 RDI: ffff9d4713f9c820
RBP: ffffb15fc915bc60 R08: ffff9d4713f9c828 R09: ffff9d4713f9c828
R10: 0000000000000228 R11: ffffb15fc915bcf0 R12: ffff9d4713f9c820
R13: 0000000000000004 R14: ffff9d47305a9980 R15: 0000000000000000
FS:  0000000000000000(0000) GS:ffff9d4643c00000(0000) knlGS:0000000000000000
CS:  0010 DS: 0000 ES: 0000 CR0: 0000000080050033
CR2: 00007fd63a1b6008 CR3: 0000000288bd6003 CR4: 0000000000370ef0
Call Trace:
 <IRQ>
 ? show_trace_log_lvl+0x1d6/0x2ea
 ? show_trace_log_lvl+0x1d6/0x2ea
 ? add_wait_queue+0x6b/0x80
 ? show_regs.part.0+0x23/0x29
 ? show_regs.cold+0x8/0xd
 ? watchdog_timer_fn+0x1be/0x220
 ? lockup_detector_update_enable+0x60/0x60
 ? __hrtimer_run_queues+0x107/0x230
 ? read_hv_clock_tsc_cs+0x9/0x30
 ? hrtimer_interrupt+0x101/0x220
 ? hv_stimer0_isr+0x20/0x30
 ? __sysvec_hyperv_stimer0+0x32/0x70
 ? sysvec_hyperv_stimer0+0x7b/0x90
 </IRQ>
 <TASK>
 ? asm_sysvec_hyperv_stimer0+0x1b/0x20
 ? _raw_spin_unlock_irqrestore+0x25/0x30
 add_wait_queue+0x6b/0x80
 do_wait+0x52/0x310
 kernel_wait4+0xaf/0x150
 ? thread_group_exited+0x50/0x50
 zap_pid_ns_processes+0x111/0x1a0
 forget_original_parent+0x348/0x360
 exit_notify+0x4a/0x210
 do_exit+0x24f/0x3c0
 do_group_exit+0x3b/0xb0
 __x64_sys_exit_group+0x18/0x20
 x64_sys_call+0x1937/0x1fa0
 do_syscall_64+0x56/0xb0
 ? do_user_addr_fault+0x1e7/0x670
 ? exit_to_user_mode_prepare+0x37/0xb0
 ? irqentry_exit_to_user_mode+0x17/0x20
 ? irqentry_exit+0x1d/0x30
 ? exc_page_fault+0x89/0x170
 entry_SYSCALL_64_after_hwframe+0x67/0xd1
RIP: 0033:0x7f60019daf8e
Code: Unable to access opcode bytes at RIP 0x7f60019daf64.
RSP: 002b:00007fff2812a468 EFLAGS: 00000246 ORIG_RAX: 00000000000000e7
RAX: ffffffffffffffda RBX: 00007f5ffeda01b0 RCX: 00007f60019daf8e
RDX: 00007f6001a560c0 RSI: 0000000000000000 RDI: 0000000000000001
RBP: 00007fff2812a4b0 R08: 0000000000000024 R09: 0000000800000000
R10: 0000000000000003 R11: 0000000000000246 R12: 0000000000000001
R13: 00007f60016f4a90 R14: 0000000000000000 R15: 00007f5ffede4d50
 </TASK>
```
Note zombie threads
```
root@ubuntu:/home/rachel# ps aux | grep Z
USER         PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root        5291  0.0  0.0      0     0 pts/2    Zl+  23:33   0:00 [rcudeadlock] <defunct>
root        5307 99.9  0.0      0     0 pts/2    Zl+  23:33  12:34 [rcudeadlock] <defunct>
root        5449  0.0  0.0      0     0 pts/2    Zl+  23:33   0:00 [rcudeadlock] <defunct>
root        5659  0.0  0.0      0     0 pts/2    Z+   23:33   0:00 [bash] <defunct>
root        5682  0.0  0.0      0     0 pts/2    Z+   23:33   0:00 [bash] <defunct>
root       11844  0.0  0.0   6480  2272 pts/3    R+   23:46   0:00 grep --color=auto Z
```
"5449" zombie process shows a thread is stuck on synchronize_rcu_expedited
```
root@ubuntu:/home/rachel# ls /proc/5449/task
5449  5565
root@ubuntu:/home/rachel# cat /proc/5565/stack
[<0>] synchronize_rcu_expedited+0x120/0x1b0
[<0>] namespace_unlock+0xd6/0x1b0
[<0>] put_mnt_ns+0x74/0xa0
[<0>] free_nsproxy+0x1c/0x1b0
[<0>] switch_task_namespaces+0x5e/0x70
[<0>] exit_task_namespaces+0x10/0x20
[<0>] do_exit+0x212/0x3c0
[<0>] io_sq_thread+0x457/0x5b0
[<0>] ret_from_fork+0x22/0x30 
```

# Usage
## npm repro
prerequisite: Install docker

Get docker image
```
# get image so that script doesn't keep pulling for it
sudo docker run telescope.azurecr.io/issue-repro/zombie:v1.1.11
```
Run script
```
./rcu-npm-repro.sh
```
This script creates several containers. Each container runs in new pid and mount namespaces. The container's entrypoint is `npm run task && npm start`.
- npm run task: This command is to run `npm run zombie & npm run done` command.
- npm run zombie: It's to run `while true; do echo zombie; sleep 1; done`. Infinite loop to print zombies.
- npm run done: It's to run `echo done`. Short live process.
- npm start: It's also a short live process. It will exit in a few seconds.

When `npm start` exits, the process tree in that pid namespace will be like
```
npm start (pid 1)
   |__npm run zombie
           |__ sh -c "whle true; do echo zombie; sleep 1; done"
```
## golang repro
```
go mod init rcudeadlock.go
go mod tidy

CGO_ENABLED=0 go build -o ./rcudeadlock ./
sudo ./rcudeadlock
```
This golang program is to simulate the npm reproducer without involving docker as dependency. This binary is using re-exec self to support multiple subcommands. It  also sets up processes in new pid and mount namespaces by unshare, since the `put_mnt_ns` is a critical code path in the kernel to reproduce this issue. Both mount and pid namespaces are required in this issue.

The entrypoint of new pid and mount namespaces is `rcudeadlock task && rcudeadlock start`.
- rcudeadlock task: This command is to run `rcudeadlock zombie & rcudeadlock done`
- rcudeadlock zombie: It's to run `bash -c "while true; do echo zombie; sleep 1; done"`. Infinite loop to print zombies.
- rcudeadlock done: Prints done and exits.
- rcudeadlock start: Prints `AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA` 10 times and exits.

When `rcudeadlock start` exits, the process tree in that pid namespace will be like
```
rcudeadlock start (pid 1)
   |__rcudeadlock zombie
           |__bash -c "while true; do echo zombie; sleep 1; done".
```

Each rcudeadlock process will set up 4 idle io_uring threads before handling commands, like `task`, `zombie`, `done` and `start`. That is similar to npm reproducer. Not sure that it's related to io_uring. But with io_uring idle threads, it's easy to reproduce this issue.
