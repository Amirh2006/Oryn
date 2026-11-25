import React from 'react';

export const MediaCardSkeleton = () => {
  return (
    <div className="flex flex-col rounded-xl overflow-hidden bg-gray-100 dark:bg-cinema-850 border border-light-border dark:border-cinema-border animate-pulse">
      {/* Poster Aspect Ratio 2:3 */}
      <div className="w-full aspect-[2/3] bg-gray-200 dark:bg-cinema-800" />
      
      {/* Card Info */}
      <div className="p-3 space-y-2">
        <div className="h-4 bg-gray-200 dark:bg-cinema-800 rounded w-3/4" />
        <div className="flex items-center justify-between">
          <div className="h-3 bg-gray-200 dark:bg-cinema-800 rounded w-1/4" />
          <div className="h-3 bg-gray-200 dark:bg-cinema-800 rounded w-1/4" />
        </div>
      </div>
    </div>
  );
};

export const HeroBannerSkeleton = () => {
  return (
    <div className="relative rounded-2xl overflow-hidden bg-gray-100 dark:bg-cinema-900 border border-light-border dark:border-cinema-border p-8 sm:p-12 animate-pulse">
      <div className="max-w-2xl space-y-4">
        <div className="h-6 w-32 bg-gray-200 dark:bg-cinema-800 rounded-full" />
        <div className="h-10 w-4/5 bg-gray-200 dark:bg-cinema-800 rounded-lg" />
        <div className="h-4 w-full bg-gray-200 dark:bg-cinema-800 rounded" />
        <div className="h-4 w-2/3 bg-gray-200 dark:bg-cinema-800 rounded" />
        <div className="pt-4 flex items-center gap-3">
          <div className="h-10 w-36 bg-gray-200 dark:bg-cinema-800 rounded-lg" />
          <div className="h-10 w-36 bg-gray-200 dark:bg-cinema-800 rounded-lg" />
        </div>
      </div>
    </div>
  );
};

export const SectionSkeleton = ({ title, count = 4 }) => {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="h-6 w-48 bg-gray-200 dark:bg-cinema-800 rounded" />
        <div className="h-4 w-16 bg-gray-200 dark:bg-cinema-800 rounded" />
      </div>
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
        {Array.from({ length: count }).map((_, i) => (
          <MediaCardSkeleton key={i} />
        ))}
      </div>
    </div>
  );
};
