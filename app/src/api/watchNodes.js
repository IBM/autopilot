export default async function watchNodes() {
    const endpoint = import.meta.env.VITE_KUBERNETES_ENDPOINT;
    const apiUrl = `${endpoint}/api/v1/nodes?watch=true`;

    try {
        const response = await fetch(apiUrl);
        const reader = response.body.getReader();
        const utf8Decoder = new TextDecoder('utf-8');

        let result = '';
        let nodesList = [];

        while (true) {
            const { done, value } = await reader.read();
            if (done) {
                break;
            }

            result += utf8Decoder.decode(value, { stream: true });

            const events = result.split('\n');
            result = events.pop();  // Keeping the last incomplete line for the next iteration

            events.forEach(event => {
                if (event.trim()) {
                    const parsedEvent = JSON.parse(event);

                    if (parsedEvent.type === "ADDED" || parsedEvent.type === "MODIFIED") {
                        const nodeName = parsedEvent.object.metadata.name;
                        if (!nodesList.includes(nodeName)) {
                            nodesList.push(nodeName);
                            console.log(`Node added or modified: ${nodeName}`);
                            console.log('Node list:', nodesList);
                        }
                    } else if (parsedEvent.type === "DELETED") {
                        const nodeName = parsedEvent.object.metadata.name;
                        nodesList = nodesList.filter(name => name !== nodeName);
                        console.log(`Node deleted: ${nodeName}`);
                    }
                }
            });
        }
        return nodesList;
    } catch (error) {
        console.error('Error watching nodes:', error);
        throw error;
    }
}