import React from 'react';
import { Search } from '@carbon/react';

// SearchInput component with label prop
function SearchInput({ searchQuery, setSearchQuery, label = "Search" }) {
    return (
        <Search
            id="search-node-name"
            labelText={label} 
            placeholder={label}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)} // Update search query
            size="md" 
        />
    );
}

export default SearchInput;
