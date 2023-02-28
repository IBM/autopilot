#include <stdio.h>
#include <stdlib.h>
#include <math.h>
#include <unistd.h>
#include <sys/time.h>
#include <cuda_runtime.h>
#include <cublas_v2.h>
#include <nvml.h>

#define MAX_BLOCKS 512
#define THREADS_PER_BLOCK 256
#define btoa(x) ((x)?"true":"false")

double cuda_dgemm(const char *, const char *, int *, int *, int *, double *, double *, int *, double *, int *, double *, double *, int *);
void cuda_dgemm_free();

#define CUDA_RC(rc) if( (rc) != cudaSuccess ) \
  {printf("Error %s at %s line %d\n", cudaGetErrorString(cudaGetLastError()), __FILE__,__LINE__); exit(1);}

#define CUDA_CHECK()  if( (cudaPeekAtLastError()) != cudaSuccess )        \
  {printf("Error %s at %s line %d\n", cudaGetErrorString(cudaGetLastError()), __FILE__,__LINE__-1); exit(1);}

double walltime(void);

__global__ void daxpy(const double alpha, const double * x, double * y, int npts) 
{
   for (int i = blockDim.x * blockIdx.x + threadIdx.x;  i < npts; i += blockDim.x * gridDim.x) y[i] = alpha*x[i] + y[i];
}

static nvmlDevice_t nvmldevice;
static unsigned int temperature, power, smMHz;

int main(int argc, char * argv[])
{
   int i, d, npts, iter, maxiter, mydevice, numDevices;
   double * __restrict__ x, * __restrict__ y;
   double * dev_x, * dev_y;
   double * Amat, * Bmat, * Cmat;
   int m, n, k, lda, ldb, ldc;
   double alpha, beta;
   double  BW_pinned_h2d, BW_pageable_h2d, BW_pinned_d2h, BW_pageable_d2h, BW_daxpy, TFlops;
   double time1, time2;
   cudaDeviceProp prop;
   double * metrics;
   nvmlDevice_t *device;
   unsigned int device_count;

   npts = 1024*1024*(1024/8);


   // initialize nvml
   if (NVML_SUCCESS != nvmlInit()) {
      fprintf(stderr, "failed to initialize NVML ... exiting\n");
   }   

   if (NVML_SUCCESS != nvmlDeviceGetCount(&device_count)) {
      fprintf(stderr, "nvmlDeviceGetCount failed ... exiting\n");
   }   

   device = (nvmlDevice_t *) malloc(device_count*sizeof(nvmlDevice_t));

   for (i = 0; i < device_count; i++) {
      if (NVML_SUCCESS != nvmlDeviceGetHandleByIndex(i, &device[i])) {
         fprintf(stderr, "nvmlDeviceGetHandleByIndex failed ... exiting\n");
      }   
   }

   // set matrix dimensions large enough to reach close to peak Flops
   m = 8192; n = 8192; k = 8192;
   Amat = (double *) malloc(m*k*sizeof(double));
   Bmat = (double *) malloc(k*n*sizeof(double));
   Cmat = (double *) malloc(m*n*sizeof(double));

#pragma omp parallel for
   for (i=0; i<(m*k); i++) Amat[i] = 1.2e-2*((double) (i%100));
#pragma omp parallel for
   for (i=0; i<(k*n); i++) Bmat[i] = 1.5e-3*((double) ((i + 100)%1000));
#pragma omp parallel for
   for (i=0; i<(m*n); i++) Cmat[i] = 1.5e-3*((double) ((i + 500)%1000));

   CUDA_RC(cudaGetDeviceCount(&numDevices));


   metrics = (double *) malloc(numDevices*9*sizeof(double));
   y = (double *) malloc(npts*sizeof(double));
   
   bool* faulty = (bool*) malloc(numDevices*sizeof(bool));
   for (i = 0; i < numDevices; ++i)
      faulty[i] = false;


   for (d = 0; d < numDevices; d++) {
      mydevice = d; /*local_rank % numDevices;*/

      // assign nvmldevice to this rank's GPU
      nvmldevice = device[mydevice];

         CUDA_RC(cudaSetDevice(mydevice));
         CUDA_RC(cudaGetDeviceProperties(&prop, mydevice));
      
         // use pinned memory for x, pageable memory for y
         CUDA_RC(cudaMallocHost((void **)&x, npts*sizeof(double)));
         //   y = (double *) malloc(npts*sizeof(double));

         CUDA_RC(cudaMalloc((void **)&dev_x, npts*sizeof(double)));
         CUDA_RC(cudaMalloc((void **)&dev_y, npts*sizeof(double)));

         #pragma omp parallel for
         for (i=0; i<npts; i++) x[i] = (double) (i%10);
         #pragma omp parallel for
         for (i=0; i<npts; i++) y[i] = (double) (i%100);

         alpha = 3.0;
         maxiter = 20;

         

         time1 = walltime();
         CUDA_RC(cudaMemcpy(dev_x, x, npts*sizeof(double), cudaMemcpyHostToDevice));
         CUDA_RC(cudaDeviceSynchronize());
         time2 = walltime();

         BW_pinned_h2d = 8.0e-9*((double) npts)/(time2 - time1);
         metrics[9*d+0] = BW_pinned_h2d;
         // Check here for low values in pinned h2d
         if (BW_pinned_h2d < 4)
            faulty[d] = true;

         time1 = walltime();
         CUDA_RC(cudaMemcpy(dev_y, y, npts*sizeof(double), cudaMemcpyHostToDevice));
         CUDA_RC(cudaDeviceSynchronize());
         time2 = walltime();

         BW_pageable_h2d = 8.0e-9*((double) npts)/(time2 - time1);
         metrics[9*d+1] = BW_pageable_h2d;
   
         time1 = walltime();
         CUDA_RC(cudaMemcpy(x, dev_x, npts*sizeof(double), cudaMemcpyDeviceToHost));
         CUDA_RC(cudaDeviceSynchronize());
         time2 = walltime();

         BW_pinned_d2h = 8.0e-9*((double) npts)/(time2 - time1);
         metrics[9*d+2] = BW_pinned_d2h;
         

         time1 = walltime();
         CUDA_RC(cudaMemcpy(y, dev_y, npts*sizeof(double), cudaMemcpyDeviceToHost));
         CUDA_RC(cudaDeviceSynchronize());
         time2 = walltime();

         BW_pageable_d2h = 8.0e-9*((double) npts)/(time2 - time1);
         metrics[9*d+3] = BW_pageable_d2h;

         int threadsPerBlock = THREADS_PER_BLOCK;
         int numBlocks = (npts + threadsPerBlock - 1) / threadsPerBlock;
         if (numBlocks > MAX_BLOCKS) numBlocks = MAX_BLOCKS;

         time1 = walltime();
         for (iter=0; iter<maxiter; iter++) {
            daxpy<<<numBlocks, threadsPerBlock>>>(alpha, dev_x, dev_y, npts);
            CUDA_CHECK();
         }
         CUDA_RC(cudaDeviceSynchronize());
         time2 = walltime();

         BW_daxpy = 3.0*8.0e-9*((double) npts)*((double) maxiter)/(time2 - time1);
         metrics[9*d+4] = BW_daxpy;
         if(BW_daxpy < 1300)
            faulty[d] = true;

         //   free(y);
         CUDA_RC(cudaFreeHost(x));
         CUDA_RC(cudaFree(dev_x));
         CUDA_RC(cudaFree(dev_y));

         beta = 0.0; lda = m; ldb = k; ldc = m;
         TFlops = cuda_dgemm("N", "N", &m, &n, &k, &alpha, Amat, &lda, Bmat, &ldb, &beta, Cmat, &ldc);
         cuda_dgemm_free();
         metrics[9*d+5] = TFlops;
         if(TFlops < 16)
            faulty[d] = true;

         metrics[9*d+6] = (double) temperature;
         metrics[9*d+7] = 1.0e-3*((double) power);  // convert to Watts
         metrics[9*d+8] = (double) smMHz;
   }
   printf(" GPU H2D(p)  H2D   D2H(p)  D2H   daxpy  dgemm   temp     power     smMHz\n");
   for (d = 0; d < numDevices; d++) {    
      printf("%3d %6.2lf %6.2lf %6.2lf %6.2lf %7.2lf %6.2lf %6.0lf %8.0lf %8.0lf\n", 
               d, metrics[9*d], metrics[9*d+1], metrics[9*d+2], metrics[9*d+3], metrics[9*d+4], metrics[9*d+5], metrics[9*d+6], metrics[9*d+7], metrics[9*d+8]);
   }
   printf("Summary of GPU errors:");
   bool allgood = true;
   for (d = 0; d < numDevices; d++) {
      if (faulty[d]) {
         allgood = false;
         printf("GPU %d -- H2D(p): %f; daxpy: %f; dgemm: %f", d, metrics[9*d+0], metrics[9*d+4], metrics[9*d+5]);
      }
   }
   if (allgood) {
      printf(" NONE ");
   }
   free(y);
   free(metrics);
   free(faulty);
   return 0;
}

double walltime(void)
{
  double elapsed;
  struct timeval tv;
  gettimeofday(&tv,NULL);
  elapsed = ((double) tv.tv_sec) + 1.0e-6*((double) tv.tv_usec);
  return elapsed;
}


// variables for cublas dgemm wrapper
static double * d_A, * d_B, * d_C;
static cublasHandle_t handle;

// use the Fortran dgemm argument list
double cuda_dgemm(const char * transa, const char * transb, int * m, int * n, int * k, 
                  double * alpha, double * A, int * lda, double * B, int * ldb, 
                  double * beta, double * C, int * ldc)
{
   int M, N, K, LDA, LDB, LDC;
   int asize, bsize, csize;
   double time1, time2, TFlops;
   cublasOperation_t opA, opB;
   int iter, maxiter = 400, sample_iter = 350;

   M = *m; N = *n; K = *k;
   LDA = *lda; LDB = *ldb; LDC = *ldc;

   asize = M*K;
   bsize = K*N;
   csize = M*N;

   cublasCreate(&handle);
   cudaMalloc((void **)&d_A, asize*sizeof(double));
   cudaMalloc((void **)&d_B, bsize*sizeof(double));
   cudaMalloc((void **)&d_C, csize*sizeof(double));

   cublasSetVector(asize, sizeof(double), A, 1, d_A, 1);
   cublasSetVector(bsize, sizeof(double), B, 1, d_B, 1);
   cublasSetVector(csize, sizeof(double), C, 1, d_C, 1);

   if      (transa[0] == 'n' || transa[0] == 'N') opA = CUBLAS_OP_N;
   else if (transa[0] == 't' || transa[0] == 'T') opA = CUBLAS_OP_T;

   if      (transb[0] == 'n' || transb[0] == 'N') opB = CUBLAS_OP_N;
   else if (transb[0] == 't' || transb[0] == 'T') opB = CUBLAS_OP_T;


   // call one time outside the timers, then time it
   cublasDgemm(handle, opA, opB, M, N, K, alpha, d_A, LDA, d_B, LDB, beta, d_C, LDC);
   cudaDeviceSynchronize();

   time1 = walltime();
   for (iter = 0; iter < maxiter; iter++) {
      cublasDgemm(handle, opA, opB, M, N, K, alpha, d_A, LDA, d_B, LDB, beta, d_C, LDC);
      if (iter == sample_iter) {
         if (NVML_SUCCESS != nvmlDeviceGetTemperature(nvmldevice, NVML_TEMPERATURE_GPU, &temperature)) temperature = 0; 
         if (NVML_SUCCESS != nvmlDeviceGetPowerUsage(nvmldevice, &power)) power = 0;
         if (NVML_SUCCESS != nvmlDeviceGetClockInfo(nvmldevice, NVML_CLOCK_SM, &smMHz)) smMHz = 0;
      }
      cudaDeviceSynchronize();
   }
   time2 = walltime();
   TFlops = 2.0e-12*((double) maxiter)*((double) M)*((double) N)*((double) K)/(time2 - time1);

   cudaMemcpy(C, d_C, csize*sizeof(double), cudaMemcpyDeviceToHost);

   return TFlops;
}

void cuda_dgemm_free()
{
   cudaFree(d_A);
   cudaFree(d_B);
   cudaFree(d_C);
   cublasDestroy(handle);
   return;
}
