import React from 'react';
import Button from './components/Button';

function Testing() {

    const runTests = () => {
        console.log('Run Tests clicked');
    };

    const selectAllNodes = () => {
        console.log('Select All clicked');
    };

    return (
        <div>
            <h1>Run Tests</h1>

            <Button
                text="Select All"
                color="green"
                onClick={selectAllNodes}
            />

            <Button
                text="Run Tests"
                color="blue"
                onClick={runTests}
            />
        </div>
    );
}

export default Testing;
