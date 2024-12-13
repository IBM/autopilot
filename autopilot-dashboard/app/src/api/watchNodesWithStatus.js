// Referenced from https://learnk8s.io/real-time-dashboard
// Tracking events with resource version
let resourceVersion = null;

// Callbacks (onNodeChange) are used to handle changes in node names incrementally
export default async function watchNodesWithStatus(onNodeChange) {
    const endpoint = import.meta.env.VITE_KUBERNETES_ENDPOINT;
    const token = import.meta.env.VITE_SERVICE_ACC_TOKEN;

    async function startWatching() {
        // Helper fn. to reconnect watch
        function reconnect() {
            startWatching();
        }

        try {
            const apiUrl = `${endpoint}/api/v1/nodes?watch=true${resourceVersion
                ? `&resourceVersion=${resourceVersion}` : ''}`;

            const headers = {};

            if (token) {
                headers['Authorization'] = `Bearer ${token}`;
                headers['Content-Type'] = 'application/json';
            }

            const response = await fetch(apiUrl, { headers });

            // If resource version is stale, do full refresh
            if (!response.ok) {
                if (response.status === 401) {
                    console.warn(`ResourceVersion stale, doing full refresh...`);
                    resourceVersion = null; // Resetting for a full refresh
                }
                console.error(`Failed to connect: ${response.statusText}`);
                return;
            }

            const reader = response.body.getReader();
            const utf8Decoder = new TextDecoder('utf-8');
            let buffer = '';

            async function readStream() {
                const { done, value } = await reader.read();
                if (done) {
                    reconnect();
                    return;
                }

                buffer += utf8Decoder.decode(value, { stream: true });
                buffer = processBuffer(buffer, onNodeChange);

                await readStream();
            }

            await readStream();
        } catch (error) {
            console.error('Error watching nodes:', error);
            setTimeout(reconnect, 5000); // Non-blocking retry
        }
    }

    await startWatching();
}

// Helper fn to process buffer
// Pass node updates via the callback
function processBuffer(buffer, onNodeChange) {
    const remainingBuffer = findLine(buffer, (line) => {
        try {
            const event = JSON.parse(line);
            const node = event.object;

            resourceVersion = node.metadata.resourceVersion;

            const nodeName = node.metadata.name;

            if (event.type === "ADDED" || event.type === "MODIFIED") {
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
                let dcgmDetails = [];

                // ERR_Year-Month-Date_Hour.Minute.UTC_Diagnostic_One.gpuNumber,Diagnostic_Two.gpuNumber
                // Example: ERR_2024-10-10_19.12.03UTC_page_retirement_row_remap.0
                if (dcgmLevel3Label.startsWith('ERR')) {
                    const [status, date, timeUTC, ...details] = dcgmLevel3Label.split('_');

                    dcgmStatus = status;
                    dcgmTimestamp = `${date} ${timeUTC.replace('UTC', ' UTC')}`;
                    dcgmDetails = details.join('_').split(",").map(detail => {
                        const [testName, gpuID] = detail.split('.');
                        return { testName, gpuID };
                    });
                } else if (dcgmLevel3Label.startsWith('PASS')) {
                    const [status, date, timeUTC] = dcgmLevel3Label.split('_');

                    dcgmStatus = status;
                    dcgmTimestamp = `${date} ${timeUTC.replace('UTC', ' UTC')}`;
                    dcgmDetails = `Pass All Tests`;
                }

                const capacity = node.status.capacity || {};
                const allocatable = node.status.allocatable || {};

                const detailedNodeInfo = {
                    name: nodeName,
                    status,
                    version,
                    architecture,
                    gpuPresent,
                    gpuHealth,
                    gpuCount,
                    gpuModel,
                    dcgmStatus,
                    dcgmTimestamp,
                    dcgmDetails,
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

                onNodeChange(detailedNodeInfo);
            } else if (event.type === "DELETED") {
                onNodeChange({ name: nodeName }, true);
            }
        } catch (error) {
            console.error('Error while parsing line:', line, '\n', error);
        }
    });

    return remainingBuffer;
}

// Helper fn to find lines in the buffer and execute a callback
function findLine(buffer, fn) {
    const newLineIndex = buffer.indexOf('\n');
    if (newLineIndex === -1) {
        return buffer; // When no new line found, return the current buffer
    }

    const chunk = buffer.slice(0, newLineIndex); // Extracting line
    const newBuffer = buffer.slice(newLineIndex + 1); // Remaining buffer

    // Processing the chunk
    fn(chunk);

    // Continue searching for more lines in the new buffer
    return findLine(newBuffer, fn);
}