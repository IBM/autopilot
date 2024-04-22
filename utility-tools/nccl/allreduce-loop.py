import os
import sys
import torch
import torch.distributed as dist
import time
import argparse

parser = argparse.ArgumentParser()
parser.add_argument("-m", "--multiplier", type=int, default=1)

args = parser.parse_args()
multiplier = args.multiplier

local_rank = int(os.environ["LOCAL_RANK"])
rank = int(os.environ["RANK"])
world_size = int(os.environ["WORLD_SIZE"])

torch.cuda.set_device(local_rank)

dist.init_process_group("nccl")

if rank == 0:
    print("NCCL version : ", torch.cuda.nccl.version(), file=sys.stderr)

torch.manual_seed(1235911);

nMB = 10000.0
npts = int(nMB*1.0e6/4.0)

Tensor = torch.rand(npts, device='cuda')
torch.cuda.synchronize()

if rank == 0:
    print("size(MB)   avgbw(GB/sec)   maxbw(GB/sec)   minbw(GB/sec)", file=sys.stderr)

for nMB in [8.0,10.0,12.5,15.0,20.0,31.6,40.0,50.0,64.0,80.0,100.0,125.0,160.0,200.0,250.0,316.0,400.0,500.0,640.0,800.0,1000.0,1250.0,1600.0,2000.0,2500.0,3160.0,4000.0,5000.0,6400.0,8000.0, 10000.0]:

    if nMB < 512.0:
        maxiter = 20*multiplier
    elif nMB < 2000.0:
        maxiter = 10*multiplier
    else:
        maxiter = 5*multiplier

    npts = int(nMB*1.0e6/4.0)
    nm1 = int(npts - 1)

    # launch two calls outside the timing loop
    dist.all_reduce(Tensor[0:nm1], op=dist.ReduceOp.SUM)
    torch.cuda.synchronize()
    dist.all_reduce(Tensor[0:nm1], op=dist.ReduceOp.SUM)
    torch.cuda.synchronize()

    tbeg = time.perf_counter()
    t1 = tbeg
    tmin = 1.0e30
    tmax = 0.0

    for i in range(maxiter):
        dist.all_reduce(Tensor[0:nm1], op=dist.ReduceOp.SUM)
        torch.cuda.synchronize()
        t2 = time.perf_counter()
        if (t2 - t1) < tmin:
            tmin = (t2 - t1)
        if (t2 - t1) > tmax:
            tmax = (t2 - t1)
        t1 = t2

    torch.cuda.synchronize()
    tend = time.perf_counter()

    elapsed = tend - tbeg

    avg_bandwidth = 4.0*2.0e-9*maxiter*npts*((world_size - 1)/world_size)/elapsed
    max_bandwidth = 4.0*2.0e-9*npts*((world_size - 1)/world_size)/tmin
    min_bandwidth = 4.0*2.0e-9*npts*((world_size - 1)/world_size)/tmax

    if rank == 0:
        print("{:7.1f}".format(nMB), "    ", "{:6.1f}".format(avg_bandwidth), "       ", "{:6.1f}".format(max_bandwidth), "        ", "{:6.1f}".format(min_bandwidth), file=sys.stderr)

