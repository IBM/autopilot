export default async function runTests(selectedTests, selectedNodes, batchParam = null) {
    try {
        const testsParam = selectedTests.join(',');
        const nodesParam = selectedNodes.join(',');

        const endpoint = import.meta.env.VITE_AUTOPILOT_ENDPOINT;
        const url = batchParam != null
            ? `${endpoint}/status?check=${testsParam}&host=${nodesParam}&batch=${batchParam}`
            : `${endpoint}/status?check=${testsParam}&host=${nodesParam}`;

        const response = await fetch(url);

        if (!response.ok) {
            throw new Error(`Error Status: ${response.status}`);
        }

        const results = await response.text();
        return results;

    } catch (error) {
        console.error('Error deploying tests:', error);
        throw error;
    }
}
