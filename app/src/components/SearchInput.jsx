import React from 'react';
import TextField from '@mui/material/TextField'; // MUI TextField for search input

// SearchInput component
function SearchInput({ searchQuery, setSearchQuery }) {
    return (
        <TextField
            label="Search by Node Name"
            variant="outlined"
            fullWidth
            margin="normal"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)} // Update search query
        />
    );
}

export default SearchInput;
