import React, { useState, useEffect } from 'react';
import { Settings } from 'lucide-react';

export const QualitySelector = ({ player }) => {
  const [qualities, setQualities] = useState([]);
  const [currentQuality, setCurrentQuality] = useState(null);
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    if (!player) return;

    const qualityLevels = player.qualityLevels && player.qualityLevels();
    if (!qualityLevels) return;

    const updateQualities = () => {
      const levels = [];
      for (let i = 0; i < qualityLevels.length; i++) {
        levels.push({
          index: i,
          height: qualityLevels[i].height,
          bitrate: qualityLevels[i].bitrate,
          enabled: qualityLevels[i].enabled,
        });
      }
      setQualities(levels);

      const enabledCount = levels.filter(q => q.enabled).length;
      if (enabledCount === levels.length) {
        setCurrentQuality({ height: 'Auto', index: -1 });
      } else {
        const current = levels.find(q => q.enabled);
        setCurrentQuality(current || null);
      }
    };

    qualityLevels.on('change', updateQualities);
    updateQualities();

    return () => {
      qualityLevels.off('change', updateQualities);
    };
  }, [player]);

  const selectQuality = (index) => {
    const qualityLevels = player.qualityLevels();
    
    if (index === -1) {
      for (let i = 0; i < qualityLevels.length; i++) {
        qualityLevels[i].enabled = true;
      }
      setCurrentQuality({ height: 'Auto', index: -1 });
    } else {
      for (let i = 0; i < qualityLevels.length; i++) {
        qualityLevels[i].enabled = i === index;
      }
      const selected = qualities.find(q => q.index === index);
      setCurrentQuality(selected);
    }
    
    setIsOpen(false);
  };

  useEffect(() => {
    const handleClickOutside = (event) => {
      if (isOpen && !event.target.closest('.quality-selector-wrapper')) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen]);

  if (qualities.length <= 1) return null;

  return (
    <div className="quality-selector-wrapper absolute bottom-16 right-4 z-20">
      <div className="relative">
        <button
          onClick={() => setIsOpen(!isOpen)}
          className="flex items-center gap-2 px-3 py-2 bg-gray-900/90 hover:bg-gray-800/95 text-white rounded-lg transition-all duration-200 backdrop-blur-md border border-white/10 shadow-lg hover:shadow-xl"
          title="Video Quality"
        >
          <Settings size={18} className={`transition-transform duration-300 ${isOpen ? 'rotate-90' : ''}`} />
          <span className="text-sm font-medium">
            {currentQuality?.height === 'Auto' ? 'Auto' : `${currentQuality?.height || '...'}p`}
          </span>
        </button>

        {isOpen && (
          <div className="absolute bottom-full right-0 mb-2 bg-gray-900/95 backdrop-blur-lg rounded-lg shadow-2xl overflow-hidden min-w-[140px] border border-white/10 animate-slide-up">
            <button
              onClick={() => selectQuality(-1)}
              className={`w-full text-left px-4 py-2.5 text-white hover:bg-blue-600/20 transition-colors text-sm font-medium flex items-center justify-between ${
                currentQuality?.height === 'Auto' ? 'bg-blue-600/30 text-blue-400' : ''
              }`}
            >
              <span>Auto</span>
              {currentQuality?.height === 'Auto' && (
                <span className="text-blue-400">✓</span>
              )}
            </button>

            <div className="h-px bg-white/10 my-1" />

            {qualities
              .sort((a, b) => b.height - a.height)
              .map((quality) => (
                <button
                  key={quality.index}
                  onClick={() => selectQuality(quality.index)}
                  className={`w-full text-left px-4 py-2.5 text-white hover:bg-blue-600/20 transition-colors text-sm font-medium flex items-center justify-between ${
                    currentQuality?.index === quality.index && currentQuality?.height !== 'Auto'
                      ? 'bg-blue-600/30 text-blue-400'
                      : ''
                  }`}
                >
                  <span>{quality.height}p</span>
                  {currentQuality?.index === quality.index && currentQuality?.height !== 'Auto' && (
                    <span className="text-blue-400">✓</span>
                  )}
                </button>
              ))}
          </div>
        )}
      </div>
    </div>
  );
};
