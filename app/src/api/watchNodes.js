// Referenced from https://learnk8s.io/real-time-dashboard

// Callbacks (onNodeChange) are used to handle changes in node names incrementally
export default async function watchNodes(onNodeChange) {
    const endpoint = import.meta.env.VITE_KUBERNETES_ENDPOINT;
    const apiUrl = `${endpoint}/api/v1/nodes?watch=true`;

    try {
        const response = await fetch(apiUrl);
        const reader = response.body.getReader();
        const utf8Decoder = new TextDecoder('utf-8');
        let buffer = '';

        async function readStream() {
            const { done, value } = await reader.read();
            if (done) {
                console.log('Watch request terminated.');
                return;
            }

            buffer += utf8Decoder.decode(value, { stream: true });
            buffer = processBuffer(buffer, onNodeChange);

            await readStream();
        }
        await readStream();
    } catch (error) {
        console.error('Error watching nodes:', error);
        throw error;
    }
}

// Helper fn to process buffer
// Pass node updates via the callback
function processBuffer(buffer, onNodeChange) {
    const remainingBuffer = findLine(buffer, (line) => {
        try {
            const event = JSON.parse(line);
            const nodeName = event.object.metadata.name;

            if (event.type === "ADDED" || event.type === "MODIFIED") {
                console.log(`Node added or modified: ${nodeName}`);
                onNodeChange(event.object);
            } else if (event.type === "DELETED") {
                console.log(`Node deleted: ${nodeName}`);
                onNodeChange(event.object, true);
            }
        } catch (error) {
            console.error('Error while parsing line:', line, '\n', error);
        }
    });

    return remainingBuffer; // Returning remaining buffer for the next read
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