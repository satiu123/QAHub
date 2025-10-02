
import React, { useState, useEffect } from 'react';
import { useLocation, Link } from 'react-router-dom';
import axios from 'axios';
import { API_BASE_URL } from '../config/api';

const SearchResult = ({ token }) => {
    const [results, setResults] = useState([]);
    const [error, setError] = useState('');
    const location = useLocation();
    const query = new URLSearchParams(location.search).get('q');

    useEffect(() => {
        if (!query) {
            setResults([]);
            return;
        }

        const fetchResults = async () => {
            try {
                const response = await axios.get(`${API_BASE_URL}/search?q=${query}`, {
                    headers: {
                        'Authorization': `Bearer ${token}`
                    }
                });
                setResults(response.data || []);
            } catch (err) {
                setError('Failed to fetch search results.');
                console.error(err);
            }
        };

        fetchResults();
    }, [query, token]);

    if (error) {
        return <div className="alert alert-danger">{error}</div>;
    }

    return (
        <div className="container">
            <h2>Search Results for "{query}"</h2>
            {results.length > 0 ? (
                <ul className="list-group">
                    {results.map(result => (
                        <li key={result.id} className="list-group-item">
                            <Link to={`/questions/${result.id}`}>{result.title}</Link>
                            <p className="mb-1">{result.content}</p>
                            {result.authorName ? (
                                <small className="text-muted">Asked by {result.authorName}</small>
                            ) : null}
                        </li>
                    ))}
                </ul>
            ) : (
                <p>No results found.</p>
            )}
        </div>
    );
};

export default SearchResult;
