Reproduce bug appears as soft lock which is associated with "zap_pid_ns_processes" and zombie defunct processes

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
Install docker and then the zombie image
```
# get image so that script doesn't keep pulling
sudo docker run telescope.azurecr.io/issue-repro/zombie:v1.1.11
```
Run script
```
./rcu-npm-repro.sh
```
## golang repro
```
go mod init
go mod tidy
go get github.com/containerd/cgroups/v3/cgroup1
go get github.com/containerd/cgroups/v3/cgroup2
go get github.com/opencontainers/runtime-spec/specs-go
go get github.com/urfave/cli
go get golang.org/x/sys/unix 

CGO_ENABLED=0 go build -o ./rcudeadlock ./
sudo ./rcudeadlock
```
