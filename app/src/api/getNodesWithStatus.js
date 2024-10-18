import useSWR from 'swr';
import axios from 'axios';

const kubernetesEndpoint = import.meta.env.VITE_KUBERNETES_ENDPOINT;

// Helper function to do fetching for SWR
const fetcher = (url) => axios.get(url).then(res => res.data);

// Using SWR to fetch and process node data with automatic updates/revalidation
export function useNodesWithStatus() {
    if (typeof kubernetesEndpoint === 'undefined') {
        throw new Error('Kubernetes endpoint not set');
    }

    const { data, error } = useSWR(`${kubernetesEndpoint}/api/v1/nodes`, fetcher, {
        refreshInterval: 10000,  // Refresh every 10 seconds
        revalidateOnFocus: true,  // Refresh when user focuses the page
    });

    const nodes = data ? data.items.map(node => {
        const nodeName = node.metadata.name;
        const role = node.metadata.labels['node-role.kubernetes.io/master'] ? 'Control Plane' :
            node.metadata.labels['node-role.kubernetes.io/worker'] ? 'Worker' : 'Unknown';
        const statusCondition = node.status.conditions.find(cond => cond.type === 'Ready') || {};
        const status = statusCondition.status || 'Unknown';
        const version = node.status.nodeInfo.kubeletVersion || 'Unknown';
        const architecture = node.status.nodeInfo.architecture || 'Unknown';

        // GPU Info
        const gpuPresent = node.metadata.labels['nvidia.com/gpu.present'] || 'Not Present';
        const gpuCount = node.metadata.labels['nvidia.com/gpu.count'] || 'Unknown';
        const gpuModel = node.metadata.labels['nvidia.com/gpu.product'] || 'Unknown';
        const gpuHealth = node.metadata.labels['autopilot.ibm.com/gpuhealth'] || 'Not Pass';

        // DCGM diagnostics
        const dcgmLevel3Label = node.metadata.labels['autopilot.ibm.com/dcgm.level.3'] || 'Not Applicable';
        let dcgmStatus = 'Unknown';
        let dcgmTimestamp = 'Unknown';
        let dcgmDetails = 'Unknown';

        // Need to take into consideration for multiple failed tests and nodes
        if (dcgmLevel3Label.startsWith('ERR')) {
            const results = dcgmLevel3Label.split('_');
            // const failedTests = [];
            // const gpuIDs = [];

            dcgmStatus = 'ERR';
            dcgmTimestamp = results[1];
            dcgmDetails = results.slice(2).join(', ');
        } else if (dcgmLevel3Label.startsWith('PASS')) {
            const results = dcgmLevel3Label.split('_');

            dcgmStatus = 'PASS';
            dcgmTimestamp = results[1];
            dcgmDetails = `Passed all tests`;
        }

        const capacity = node.status.capacity || {};
        const allocatable = node.status.allocatable || {};

        return {
            name: nodeName,
            role: role,
            status: status,
            version: version,
            architecture: architecture,

            gpuPresent: gpuPresent,
            gpuHealth: gpuHealth,
            gpuCount: gpuCount,
            gpuModel: gpuModel,

            dcgmStatus: dcgmStatus,
            dcgmTimestamp: dcgmTimestamp,
            dcgmDetails: dcgmDetails,

            capacity: {
                gpu: capacity['nvidia.com/gpu'] || 'Unknown',
                cpu: capacity.cpu || 'Unknown',
                memory: capacity.memory || 'Unknown',
            },
            allocatable: {
                gpu: allocatable['nvidia.com/gpu'] || 'Unknown',
                cpu: allocatable.cpu || 'Unknown',
                memory: allocatable.memory || 'Unknown',
            }
        };
    }) : [];

    return { nodes, error };
}