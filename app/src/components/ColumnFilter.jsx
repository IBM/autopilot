import React, { useState, useRef, useEffect } from 'react';
import { Button, MultiSelect } from '@carbon/react';
import { Filter } from '@carbon/icons-react'; // Import the filter icon
import ReactDOM from 'react-dom';
import styled from 'styled-components';

const DropdownContainer = styled.div`
    position: absolute;
    z-index: 1000;
    background: white;
    border: 1px solid #ddd;
    border-radius: 4px;
    padding: 10px;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
`;
const MultiSelectItem = styled.div`
    white-space: nowrap; // Prevent line breaks
    overflow: hidden; // Hide overflow text
    text-overflow: ellipsis; // Add ellipsis for overflow text
    max-width: 100%; // Ensures it doesn't exceed the container
`;
const ColumnFilter = ({ label, items, selectedFilters, onFilterChange }) => {
    const [filterOpen, setFilterOpen] = useState(false);
    const buttonRef = useRef(null);
    const dropdownRef = useRef(null);
    const [dropdownPosition, setDropdownPosition] = useState({ top: 0, left: 0 });

    const handleFilterToggle = () => {
        if (filterOpen) {
            setFilterOpen(false);
        } else {
            const buttonRect = buttonRef.current.getBoundingClientRect();
            const dropdownHeight = 300; // Adjust as needed
            const dropdownWidth = 200; // Adjust as needed

            let top = buttonRect.bottom + window.scrollY + 5;
            if (top + dropdownHeight > window.innerHeight) {
                top = buttonRect.top + window.scrollY - dropdownHeight - 5;
            }

            let left = buttonRect.left + window.scrollX;
            if (left + dropdownWidth > window.innerWidth) {
                left = window.innerWidth - dropdownWidth - 10;
            }

            setDropdownPosition({ top, left });
            setFilterOpen(true);
        }
    };

    const handleFilterChange = (selectedItems) => {
        onFilterChange(selectedItems.map(item => item.id));
    };

    useEffect(() => {
        const handleClickOutside = (event) => {
            if (
                filterOpen &&
                dropdownRef.current &&
                !dropdownRef.current.contains(event.target) &&
                !buttonRef.current.contains(event.target)
            ) {
                setFilterOpen(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, [filterOpen]);

    return (
        <>
            <Button
                ref={buttonRef}
                hasIconOnly
                renderIcon={Filter} // Use the filter icon
                onClick={handleFilterToggle}
                iconDescription="Filter"
                kind="ghost"
                aria-haspopup="true"
                aria-expanded={filterOpen}
            />
            {filterOpen && ReactDOM.createPortal(
                <DropdownContainer ref={dropdownRef} style={{
                    top: dropdownPosition.top,
                    left: dropdownPosition.left,
                }}>
                    <MultiSelect
                        items={items.map(item => ({
                            id: item,
                            label: item,
                        }))}
                        itemToString={(item) => (item ? item.label : '')}
                        onChange={({ selectedItems }) => handleFilterChange(selectedItems)}
                        label={label} // Use label for the dropdown
                        placeholder={`Select ${label}`} // Adjust placeholder if needed
                        initialSelectedItems={items.filter(item => selectedFilters.includes(item)).map(item => ({
                            id: item,
                            label: item,
                        }))}
                    />
                </DropdownContainer>,
                document.body
            )}
        </>
    );
};

export default ColumnFilter;