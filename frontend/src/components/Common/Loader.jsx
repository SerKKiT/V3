import React from 'react';

export const Loader = ({ size = 'md' }) => {
  const sizes = {
    sm: 'w-4 h-4',
    md: 'w-8 h-8',
    lg: 'w-12 h-12',
  };

  return (
    <div className="flex justify-center items-center">
      <div className={`${sizes[size]} border-4 border-primary-600 border-t-transparent rounded-full animate-spin`}></div>
    </div>
  );
};
