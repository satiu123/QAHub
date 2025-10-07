
import React, { useState, useEffect } from 'react';
import { useLocation, Link } from 'react-router-dom';
import { searchQuestions } from '../services/searchService';

const SearchResult = ({ token }) => {
    const [results, setResults] = useState([]);
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);
    const location = useLocation();
    const query = new URLSearchParams(location.search).get('q');

    useEffect(() => {
        if (!query) {
            setResults([]);
            return;
        }

        const fetchResults = async () => {
            setLoading(true);
            setError('');
            try {
                const questions = await searchQuestions({
                    token,
                    query,
                    limit: 20,
                    offset: 0,
                });
                setResults(questions);
            } catch (err) {
                setError('Failed to fetch search results.');
                console.error('Search error:', err);
            } finally {
                setLoading(false);
            }
        };

        fetchResults();
    }, [query, token]);

    if (error) {
        return <div className="alert alert-danger">{error}</div>;
    }

    if (loading) {
        return (
            <div className="container">
                <div className="text-center mt-5">
                    <div className="spinner-border" role="status">
                        <span className="visually-hidden">Loading...</span>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="container">
            <h2>Search Results for "{query}"</h2>
            {results.length > 0 ? (
                <ul className="list-group">
                    {results.map(result => (
                        <li key={result.id} className="list-group-item">
                            <Link to={`/questions/${result.id}`}>
                                <h5>{result.title}</h5>
                            </Link>
                            <p className="mb-1">{result.content}</p>
                            <small className="text-muted">
                                Author ID: {result.author_id}
                                {result.created_at && ` • Created: ${new Date(result.created_at).toLocaleDateString()}`}
                            </small>
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
