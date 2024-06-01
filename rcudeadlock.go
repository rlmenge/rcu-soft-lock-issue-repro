package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/containerd/cgroups/v3"
	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/iceber/iouring-go"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/urfave/cli"
	"golang.org/x/sys/unix"
)

func main() {
	if err := Entrypoint().Run(os.Args); err != nil {
		panic(err)
	}
}

func Entrypoint() *cli.App {
	return &cli.App{
		Commands: []cli.Command{
			RCUDeadlockCommand(),
			RunTask(),
			RunZombie(),
			RunDone(),
			RunStart(),
		},
		Action: func(cliCtx *cli.Context) error {
			for {
				var wg sync.WaitGroup

				for i := 0; i < 10; i++ {
					wg.Add(1)

					cgroupPath := fmt.Sprintf("/rcu-deadlock-issue-%v", i)
					go func() {
						defer wg.Done()

						// start to re-exec
						exe, err := os.Executable()
						if err != nil {
							panic(fmt.Errorf("failed to get executable path: %w", err))
						}

						cmd := exec.Command(exe, "rcu-deadlock", "--cpu_quota_us", "1000", "--cpu_period_us", "10000", cgroupPath)
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr
						if err := cmd.Run(); err != nil {
							panic(fmt.Errorf("failed to run rcu-deadlock: %w", err))
						}
					}()
				}
				wg.Wait()
				time.Sleep(1 * time.Second)
			}
			return nil
		},
	}
}

func RCUDeadlockCommand() cli.Command {
	return cli.Command{
		Name: "rcu-deadlock",
		Flags: []cli.Flag{
			cli.Int64Flag{
				Name: "cpu_quota_us",
			},
			cli.Uint64Flag{
				Name: "cpu_period_us",
			},
		},
		Action: func(cliCtx *cli.Context) error {
			cgroupPath := cliCtx.Args().Get(0)
			if cgroupPath == "" {
				return fmt.Errorf("required cgroupPath as first arg")
			}

			quota := cliCtx.Int64("cpu_quota_us")
			period := cliCtx.Uint64("cpu_period_us")
			switch mode := cgroups.Mode(); mode {
			case cgroups.Unified:
				mgr, err := cgroup2.NewManager("/sys/fs/cgroup", cgroupPath, &cgroup2.Resources{
					CPU: &cgroup2.CPU{
						Max: cgroup2.NewCPUMax(&quota, &period),
					},
				})
				if err != nil {
					return fmt.Errorf("failed to create cgroupv2 manager: %w", err)
				}
				if err := mgr.AddProc(uint64(os.Getpid())); err != nil {
					return fmt.Errorf("failed to move to cgroup %s: %w", cgroupPath, err)
				}

			case cgroups.Legacy:
				mgr, err := cgroup1.Load(cgroup1.RootPath)
				if err != nil {
					return fmt.Errorf("failed to load cgroupv1: %w", err)
				}
				mgr, err = mgr.New(cgroupPath, &specs.LinuxResources{
					CPU: &specs.LinuxCPU{
						Quota:  &quota,
						Period: &period,
					},
				})
				if err != nil {
					return fmt.Errorf("failed to update cpu quota/period: %w", err)
				}
				if err := mgr.AddProc(uint64(os.Getpid())); err != nil {
					return fmt.Errorf("failed to move to cgroup %s: %w", cgroupPath, err)
				}
			default:
				return fmt.Errorf("only support cgroupv1 and cgroupv2, excluding hybrid mode: %v", mode)
			}

			// NOTE: Set it self as subreaper so that it can be the
			// parent of child double-forked by unshare command.
			if err := setSubreaper(1); err != nil {
				return fmt.Errorf("failed to set sub reaper: %w", err)
			}

			// start to re-exec
			exe, err := os.Executable()
			if err != nil {
				return fmt.Errorf("failed to get executable path: %w", err)
			}

			cmd := exec.Command(exe, "unshareworkload")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to run workload in pid/mnt namespace: %w", err)
			}

			var (
				ws  unix.WaitStatus
				rus unix.Rusage
			)

			pid, err := unix.Wait4(-1, &ws, 0, &rus)
			if err != nil {
				return fmt.Errorf("failed to wait workload: %w", err)
			}
			fmt.Println(pid, " Exit")
			return nil
		},
	}
}

func RunTask() cli.Command {
	return cli.Command{
		Name: "task",
		Action: func(cliCtx *cli.Context) error {
			if err := createFourIdleIOUringThreads(); err != nil {
				return err
			}

			// start to re-exec
			exe, err := os.Executable()
			if err != nil {
				return fmt.Errorf("failed to get executable path: %w", err)
			}

			cmd := exec.Command("bash", "-c", fmt.Sprintf("%s zombie & %s done", exe, exe))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to run task: %w", err)
			}
			return nil
		},
	}
}

func RunZombie() cli.Command {
	return cli.Command{
		Name: "zombie",
		Action: func(cliCtx *cli.Context) error {
			if err := createFourIdleIOUringThreads(); err != nil {
				return err
			}

			cmd := exec.Command("bash", "-c", "while true; do echo zombie; sleep 1; done")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to run zombie: %w", err)
			}
			return nil
		},
	}
}

func RunDone() cli.Command {
	return cli.Command{
		Name: "done",
		Action: func(cliCtx *cli.Context) error {
			if err := createFourIdleIOUringThreads(); err != nil {
				return err
			}

			cmd := exec.Command("echo", "done")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to run done: %w", err)
			}
			return nil
		},
	}
}

func RunStart() cli.Command {
	return cli.Command{
		Name: "start",
		Action: func(cliCtx *cli.Context) error {
			if err := createFourIdleIOUringThreads(); err != nil {
				return err
			}

			for i := 0; i < 10; i++ {
				fmt.Println("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
			}
			return nil
		},
	}
}

func init() {
	runtime.LockOSThread()

	if len(os.Args) == 2 && os.Args[1] == "unshareworkload" {
		err := unix.Unshare(unix.CLONE_NEWNS | unix.CLONE_NEWPID)
		if err != nil {
			panic(err)
		}

		// start to re-exec
		exe, err := os.Executable()
		if err != nil {
			panic(fmt.Errorf("failed to get executable path: %w", err))
		}

		cmd := exec.Command("bash", "-c", fmt.Sprintf("%s task && %s start", exe, exe))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Start()
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	}
}

// setSubreaper runs as sub-reaper
func setSubreaper(i int) error {
	return unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, uintptr(i), 0, 0, 0)
}

// createFourIdleIOUringThreads creates 4 idle iouring threads.
func createFourIdleIOUringThreads() error {
	_, err := iouring.New(64, iouring.WithSQPoll())
	if err != nil {
		return fmt.Errorf("failed to create iouring with sq_poll: %w", err)
	}

	_, err = iouring.New(256, iouring.WithSQPoll(), iouring.WithSQPollThreadIdle(10*time.Millisecond))
	if err != nil {
		return fmt.Errorf("failed to create iouring with sq_poll and sq_poll_thread_idel=10ms: %w", err)
	}

	_, err = iouring.New(64, iouring.WithSQPoll(), iouring.WithSQPollThreadIdle(10*time.Millisecond))
	if err != nil {
		return fmt.Errorf("failed to create iouring with sq_poll and sq_poll_thread_idel=10ms: %w", err)
	}

	iour, err := iouring.New(64)
	if err != nil {
		return fmt.Errorf("failed to create iouring: %w", err)
	}

	curDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working dir")
	}

	curDirFile, err := os.Open(curDir)
	if err != nil {
		return fmt.Errorf("failed to open current working dir")
	}
	defer curDirFile.Close()

	var stat unix.Statx_t
	req, err := iouring.Statx(int(curDirFile.Fd()), "./", 0, 0, &stat)
	if err != nil {
		return fmt.Errorf("failed to prepare statx iouring request: %w", err)
	}

	result, err := iour.SubmitRequest(
		req,
		make(chan iouring.Result, 1),
	)
	if err != nil {
		return fmt.Errorf("failed to submit statx request: %w", err)
	}
	<-result.Done()
	if err := result.Err(); err != nil {
		return fmt.Errorf("failed to run statx: %w", err)
	}
	fmt.Printf("DevMajor: %v, DevMinor: %v\n", stat.Dev_major, stat.Dev_minor)
	return nil
}
