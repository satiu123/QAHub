
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';

const SearchBox = () => {
    const [query, setQuery] = useState('');
    const navigate = useNavigate();

    const handleSearch = (e) => {
        e.preventDefault();
        if (query.trim()) {
            navigate(`/search?q=${query.trim()}`);
            setQuery('');
        }
    };

    return (
        <form className="d-flex" onSubmit={handleSearch}>
            <input
                className="form-control me-2"
                type="search"
                placeholder="Search questions..."
                aria-label="Search"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
            />
            <button className="btn btn-outline-success" type="submit">Search</button>
        </form>
    );
};

export default SearchBox;
