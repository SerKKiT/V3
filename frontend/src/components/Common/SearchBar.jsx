import React from 'react';
import { Search, X } from 'lucide-react';

export const SearchBar = ({ 
  value, 
  onChange, 
  onClear, 
  placeholder = 'Search...',
  className = '' 
}) => {
  return (
    <div className={`relative ${className}`}>
      <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
        <Search className="h-5 w-5 text-gray-500" />
      </div>
      
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="input-field pl-10 pr-10"
        placeholder={placeholder}
      />
      
      {value && (
        <button
          onClick={onClear}
          className="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-500 hover:text-white transition"
        >
          <X className="h-5 w-5" />
        </button>
      )}
    </div>
  );
};
