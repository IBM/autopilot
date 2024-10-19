// referenced from https://learnk8s.io/real-time-dashboard
export default async function watchNodes() {
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
            buffer = processBuffer(buffer);

            await readStream();
        }
        await readStream();
    } catch (error) {
        console.error('Error watching nodes:', error);
        throw error;
    }
}

// Helper fn to process the buffer and handle events
function processBuffer(buffer) {
    const remainingBuffer = findLine(buffer, (line) => {
        try {
            const event = JSON.parse(line);
            const nodeName = event.object.metadata.name;

            if (event.type === "ADDED" || event.type === "MODIFIED") {
                console.log(`Node added or modified: ${nodeName}`);
            } else if (event.type === "DELETED") {
                console.log(`Node deleted: ${nodeName}`);
            }
        } catch (error) {
            console.error('Error while parsing line:', line, '\n', error);
        }
    });

    return remainingBuffer; // returning remaining buffer for the next read
}

// Helper fn to find lines in the buffer and execute a callback
function findLine(buffer, fn) {
    const newLineIndex = buffer.indexOf('\n');
    if (newLineIndex === -1) {
        return buffer; // when no new line found, return the current buffer
    }

    const chunk = buffer.slice(0, newLineIndex); // extracting line
    const newBuffer = buffer.slice(newLineIndex + 1); // remaining buffer

    // processing the chunk
    fn(chunk);

    // continue searching for more lines in the new buffer
    return findLine(newBuffer, fn);
}