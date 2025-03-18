export default async function runTests(selectedTests, selectedNodes = [], jobValue = '', labelValue = '', batchValue = '', dcgmRValue = '') {
    try {
        const testsValue = selectedTests.join(',');
        const nodesValue = selectedNodes.join(',');

        let url = `/autopilot/status?check=${testsValue}`;

        if (nodesValue) {
            url += `&host=${nodesValue}`;
        }
        if (jobValue) {
            url += `&job=${jobValue}`;
        }
        if (labelValue) {
            url += `&nodelabel=${labelValue}`;
        }
        if (dcgmRValue) {
            url += `&r=${dcgmRValue}`;
        }
        if (batchValue) {
            url += `&batch=${batchValue}`;
        }

        const response = await fetch(url);

        if (!response.ok) {
            throw new Error(`Error Status: ${response.status} `);
        }

        const results = await response.text();
        return results;

    } catch (error) {
        console.error('Error deploying tests:', error);
        throw error;
    }
}
