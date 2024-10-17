import React from 'react';
import {Search} from '@carbon/react'

// SearchInput component
function SearchInput({ searchQuery, setSearchQuery }) {
    return (
        <Search
            id="search-node-name"
            labelText="Search by Node Name" // Accessible label
            placeHolder="Search by Node Name" // Placeholder text
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)} // Update search query
            size="md" 
        />
    );
}

export default SearchInput;
