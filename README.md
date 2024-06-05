Reproduce soft lock which is associated with "zap_pid_ns_processes" and zombie defunct processes

Example dmesg:
```
[Fri May 31 23:40:07 2024] watchdog: BUG: soft lockup - CPU#4 stuck for 354s! [iou-wrk-5307:5745] 
...
[Fri May 31 23:40:07 2024]  <TASK>
[Fri May 31 23:40:07 2024]  ? asm_sysvec_hyperv_stimer0+0x1b/0x20
[Fri May 31 23:40:07 2024]  ? kernel_wait4+0x9/0x150
***[Fri May 31 23:40:07 2024]  zap_pid_ns_processes+0x111/0x1a0***
[Fri May 31 23:40:07 2024]  forget_original_parent+0x348/0x360
[Fri May 31 23:40:07 2024]  exit_notify+0x4a/0x210
[Fri May 31 23:40:07 2024]  do_exit+0x24f/0x3c0
[Fri May 31 23:40:07 2024]  io_wqe_worker+0x2b3/0x320
[Fri May 31 23:40:07 2024]  ? raw_spin_rq_unlock+0x10/0x30
[Fri May 31 23:40:07 2024]  ? finish_task_switch.isra.0+0x7e/0x280
[Fri May 31 23:40:07 2024]  ? io_worker_handle_work+0x2b0/0x2b0
[Fri May 31 23:40:07 2024]  ret_from_fork+0x22/0x30
[Fri May 31 23:40:07 2024] RIP: 0033:0x0
[Fri May 31 23:40:07 2024] Code: Unable to access opcode bytes at RIP 0xffffffffffffffd6.
[Fri May 31 23:40:07 2024] RSP: 002b:0000000000000000 EFLAGS: 00000206 ORIG_RAX: 00000000000001aa
[Fri May 31 23:40:07 2024] RAX: 0000000000000000 RBX: 000000c00002e000 RCX: 00000000004a8c6a
[Fri May 31 23:40:07 2024] RDX: 0000000000000000 RSI: 0000000000000001 RDI: 000000000000000d
[Fri May 31 23:40:07 2024] RBP: 000000c000129528 R08: 0000000000000000 R09: 0000000000000000
[Fri May 31 23:40:07 2024] R10: 0000000000000000 R11: 0000000000000206 R12: 0000000000000000
[Fri May 31 23:40:07 2024] R13: 0000000000000000 R14: 000000c0000021a0 R15: 00007ff0afe8a009
[Fri May 31 23:40:07 2024]  </TASK>

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
