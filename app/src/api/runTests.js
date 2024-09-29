export default async function runTests(selectedTests, selectedNodes) {
    try {
        const testsParam = selectedTests.join(',');
        const nodesParam = selectedNodes.join(',');

        const endpoint = import.meta.env.VITE_AUTOPILOT_ENDPOINT;
        const url = `${endpoint}/status?check=${testsParam}&host=${nodesParam}`;

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
