import React, { useState } from 'react';
import Button from './components/Button';
import MultiSelect from './components/MultiSelect';

function Testing() {

    const [selectedTests, setSelectedTests] = useState([]);
    const tests = ['Test 1', 'Test 2', 'Test 3', 'Test 4'];

    const handleMultiSelectChange = (selected) => {
        setSelectedTests(selected);
    };

    const runTests = () => {
        console.log('Run Tests clicked');
        console.log(selectedTests)
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

            <MultiSelect
                options={tests}
                placeholder="Select Health Checks"
                selectedValues={selectedTests}
                handleChange={handleMultiSelectChange}
            />
        </div>
    );
}

export default Testing;
