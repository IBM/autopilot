export default async function listNodes() {
    const endpoint = import.meta.env.VITE_KUBERNETES_ENDPOINT;
    const apiUrl = `${endpoint}/api/v1/nodes`;
    const token = import.meta.env.VITE_SERVICE_ACC_TOKEN;

    try {
        const response = await fetch(apiUrl, {
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });
        const data = await response.json();

        const nodesList = data.items.map(node => node.metadata.name);
        return nodesList;
    } catch (error) {
        console.error('Error fetching nodes list:', error);
        throw error;
    }
}