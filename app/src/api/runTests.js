export default async function runTests(selectedTests, selectedNodes) {
    try {
        const testsParam = selectedTests.join(',');
        const nodesParam = selectedNodes.join(',');

        const url = `http://localhost:3333/status?check=${testsParam}&host=${nodesParam}`;

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
