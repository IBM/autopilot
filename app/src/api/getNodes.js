export default async function listNodes() {
    const endpoint = import.meta.env.VITE_KUBERNETES_ENDPOINT;
    const apiUrl = `${endpoint}/api/v1/nodes`;

    try {
        const response = await fetch(apiUrl);
        const data = await response.json();

        const nodesList = data.items.map(node => node.metadata.name);
        return nodesList;
    } catch (error) {
        console.error('Error fetching nodes list:', error);
        throw error;
    }
}