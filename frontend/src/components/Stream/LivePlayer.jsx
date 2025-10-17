import { useEffect, useRef, useState } from 'react';
import videojs from 'video.js';
import 'video.js/dist/video-js.css';
import { Play, Pause, Volume2, VolumeX, Maximize, Settings } from 'lucide-react';

export const LivePlayer = ({ hlsUrl }) => {
  const videoRef = useRef(null);
  const playerRef = useRef(null);
  const containerRef = useRef(null);
  const progressRef = useRef(null);
  
  const [isPlaying, setIsPlaying] = useState(false);
  const [volume, setVolume] = useState(1);
  const [isMuted, setIsMuted] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [seekableStart, setSeekableStart] = useState(0);
  const [seekableEnd, setSeekableEnd] = useState(0);
  const [bufferedEnd, setBufferedEnd] = useState(0);
  const [showControls, setShowControls] = useState(true);
  const [qualities, setQualities] = useState([]);
  const [currentQuality, setCurrentQuality] = useState('auto');
  const [showQualityMenu, setShowQualityMenu] = useState(false);
  const [isDragging, setIsDragging] = useState(false);
  const [seekIndicator, setSeekIndicator] = useState(null);

  let hideControlsTimeout = useRef(null);
  const hasStartedRef = useRef(false);

  useEffect(() => {
    const timer = setTimeout(() => {
      if (!videoRef.current || playerRef.current) return;

      const player = videojs(videoRef.current, {
        controls: false,
        autoplay: true,
        muted: false,
        preload: 'auto',
        fluid: true,
        liveui: false,
        html5: {
          vhs: {
            overrideNative: true,
          },
        },
      });

      playerRef.current = player;

      player.src({
        src: hlsUrl,
        type: 'application/x-mpegURL',
      });

      player.ready(() => {
        const qualityLevels = player.qualityLevels();
        if (qualityLevels) {
          qualityLevels.on('addqualitylevel', () => updateQualityList(qualityLevels));
          updateQualityList(qualityLevels);
        }

        const playPromise = player.play();
        if (playPromise !== undefined) {
          playPromise.catch(error => {
            console.log('⚠️ Autoplay blocked, trying muted:', error);
            player.muted(true);
            player.play().catch(e => console.error('❌ Failed to play:', e));
          });
        }
      });

      player.on('play', () => setIsPlaying(true));
      player.on('pause', () => setIsPlaying(false));
      player.on('volumechange', () => {
        setVolume(player.volume());
        setIsMuted(player.muted());
      });
      
      player.on('loadedmetadata', () => {
        if (hasStartedRef.current) return;
        
        const seekable = player.seekable();
        if (seekable.length > 0) {
          const end = seekable.end(0);
          const startTime = Math.max(0, end - 2);
          player.currentTime(startTime);
          hasStartedRef.current = true;
          console.log('🎬 Started from:', startTime, '(end:', end, ')');
        }
      });

      player.on('timeupdate', () => {
        if (!isDragging) {
          const time = player.currentTime();
          setCurrentTime(time);
        }
        
        const seekable = player.seekable();
        if (seekable.length > 0) {
          const start = seekable.start(0);
          const end = seekable.end(0);
          setSeekableStart(start);
          setSeekableEnd(end);
        }

        const buffered = player.buffered();
        if (buffered.length > 0) {
          const bufEnd = buffered.end(buffered.length - 1);
          setBufferedEnd(bufEnd);
        }
      });

      player.on('error', () => {
        const error = player.error();
        console.error('❌ Player error:', error);
      });
    }, 0);

    return () => {
      clearTimeout(timer);
      if (playerRef.current && !playerRef.current.isDisposed()) {
        playerRef.current.dispose();
        playerRef.current = null;
      }
    };
  }, [hlsUrl]);

  // ✅ KEYBOARD SHORTCUTS
  useEffect(() => {
    const handleKeyDown = (e) => {
      if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') {
        return;
      }

      if (!playerRef.current) return;

      switch (e.key) {
        case ' ':
        case 'k':
        case 'K':
          e.preventDefault();
          togglePlay();
          break;

        case 'ArrowLeft':
          e.preventDefault();
          const newTimeBack = Math.max(seekableStart, currentTime - 5);
          playerRef.current.currentTime(newTimeBack);
          setCurrentTime(newTimeBack);
          showSeekIndicator('-5s');
          break;

        case 'ArrowRight':
          e.preventDefault();
          const newTimeForward = Math.min(seekableEnd, currentTime + 5);
          playerRef.current.currentTime(newTimeForward);
          setCurrentTime(newTimeForward);
          showSeekIndicator('+5s');
          break;

        case 'j':
        case 'J':
          e.preventDefault();
          const newTimeJ = Math.max(seekableStart, currentTime - 10);
          playerRef.current.currentTime(newTimeJ);
          setCurrentTime(newTimeJ);
          showSeekIndicator('-10s');
          break;

        case 'l':
        case 'L':
          e.preventDefault();
          const newTimeL = Math.min(seekableEnd, currentTime + 10);
          playerRef.current.currentTime(newTimeL);
          setCurrentTime(newTimeL);
          showSeekIndicator('+10s');
          break;

        case 'ArrowUp':
          e.preventDefault();
          const newVolumeUp = Math.min(1, volume + 0.1);
          playerRef.current.volume(newVolumeUp);
          if (isMuted) playerRef.current.muted(false);
          break;

        case 'ArrowDown':
          e.preventDefault();
          const newVolumeDown = Math.max(0, volume - 0.1);
          playerRef.current.volume(newVolumeDown);
          break;

        case 'm':
        case 'M':
          e.preventDefault();
          toggleMute();
          break;

        case 'f':
        case 'F':
          e.preventDefault();
          toggleFullscreen();
          break;

        case '0':
        case '1':
        case '2':
        case '3':
        case '4':
        case '5':
        case '6':
        case '7':
        case '8':
        case '9':
          e.preventDefault();
          const percent = parseInt(e.key) * 0.1;
          const seekableDuration = seekableEnd - seekableStart;
          const newTimePercent = seekableStart + (seekableDuration * percent);
          playerRef.current.currentTime(newTimePercent);
          setCurrentTime(newTimePercent);
          break;

        case 'Home':
          e.preventDefault();
          playerRef.current.currentTime(seekableStart);
          setCurrentTime(seekableStart);
          break;

        case 'End':
          e.preventDefault();
          playerRef.current.currentTime(seekableEnd - 1);
          setCurrentTime(seekableEnd - 1);
          break;

        default:
          break;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [currentTime, seekableStart, seekableEnd, volume, isMuted, isPlaying]);

  const showSeekIndicator = (text) => {
    setSeekIndicator(text);
    setTimeout(() => setSeekIndicator(null), 500);
  };

  const updateQualityList = (qualityLevels) => {
    const options = [];
    for (let i = 0; i < qualityLevels.length; i++) {
      options.push({
        id: i,
        label: `${qualityLevels[i].height}p`,
        height: qualityLevels[i].height,
      });
    }
    options.sort((a, b) => b.height - a.height);
    setQualities(options);
  };

  const handleQualityChange = (qualityId) => {
    if (!playerRef.current) return;
    const qualityLevels = playerRef.current.qualityLevels();

    if (qualityId === 'auto') {
      for (let i = 0; i < qualityLevels.length; i++) {
        qualityLevels[i].enabled = true;
      }
      setCurrentQuality('auto');
    } else {
      for (let i = 0; i < qualityLevels.length; i++) {
        qualityLevels[i].enabled = i === qualityId;
      }
      const selected = qualities.find(q => q.id === qualityId);
      setCurrentQuality(selected?.label || 'auto');
    }
    setShowQualityMenu(false);
  };

  const togglePlay = () => {
    if (!playerRef.current) return;
    if (isPlaying) {
      playerRef.current.pause();
    } else {
      playerRef.current.play();
    }
  };

  const toggleMute = () => {
    if (!playerRef.current) return;
    playerRef.current.muted(!isMuted);
  };

  const handleVolumeChange = (e) => {
    if (!playerRef.current) return;
    const newVolume = parseFloat(e.target.value);
    playerRef.current.volume(newVolume);
    if (newVolume > 0 && isMuted) {
      playerRef.current.muted(false);
    }
  };

  const toggleFullscreen = () => {
    if (!containerRef.current) return;
    if (!document.fullscreenElement) {
      containerRef.current.requestFullscreen();
    } else {
      document.exitFullscreen();
    }
  };

  const seekToPosition = (e) => {
    if (!playerRef.current || !progressRef.current || seekableEnd === 0) return;
    
    const rect = progressRef.current.getBoundingClientRect();
    const pos = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width));
    
    const seekableDuration = seekableEnd - seekableStart;
    const newTime = seekableStart + (pos * seekableDuration);
    const clampedTime = Math.max(seekableStart, Math.min(seekableEnd, newTime));
    
    playerRef.current.currentTime(clampedTime);
    setCurrentTime(clampedTime);
  };

  const handleProgressMouseDown = (e) => {
    setIsDragging(true);
    seekToPosition(e);
  };

  const handleProgressMouseMove = (e) => {
    if (!isDragging) return;
    seekToPosition(e);
  };

  const handleProgressMouseUp = () => {
    setIsDragging(false);
  };

  useEffect(() => {
    if (isDragging) {
      window.addEventListener('mousemove', handleProgressMouseMove);
      window.addEventListener('mouseup', handleProgressMouseUp);
      return () => {
        window.removeEventListener('mousemove', handleProgressMouseMove);
        window.removeEventListener('mouseup', handleProgressMouseUp);
      };
    }
  }, [isDragging, seekableStart, seekableEnd]);

  const handleMouseMove = () => {
    setShowControls(true);
    if (hideControlsTimeout.current) {
      clearTimeout(hideControlsTimeout.current);
    }
    hideControlsTimeout.current = setTimeout(() => {
      if (isPlaying && !isDragging) {
        setShowControls(false);
      }
    }, 3000);
  };

  const formatTime = (seconds) => {
    if (!seconds || isNaN(seconds)) return '0:00';
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const getProgressPercent = () => {
    const seekableDuration = seekableEnd - seekableStart;
    if (seekableDuration <= 0) return 0;
    const progress = ((currentTime - seekableStart) / seekableDuration) * 100;
    return Math.max(0, Math.min(100, progress));
  };

  const getBufferedPercent = () => {
    const seekableDuration = seekableEnd - seekableStart;
    if (seekableDuration <= 0) return 0;
    const buffered = ((bufferedEnd - seekableStart) / seekableDuration) * 100;
    return Math.max(0, Math.min(100, buffered));
  };

  const isLive = seekableEnd > 0 && (seekableEnd - currentTime) < 5;

  return (
    <div
      ref={containerRef}
      className="relative w-full bg-black rounded-lg overflow-hidden shadow-2xl group"
      onMouseMove={handleMouseMove}
      onMouseLeave={() => isPlaying && !isDragging && setShowControls(false)}
    >
      {/* Video элемент */}
      <div data-vjs-player onClick={togglePlay} className="cursor-pointer">
        <video ref={videoRef} className="video-js w-full" />
      </div>

      {/* ✅ Seek Indicator */}
      {seekIndicator && (
        <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 bg-black/80 text-white px-6 py-3 rounded-lg text-2xl font-bold backdrop-blur-sm pointer-events-none z-40 animate-fade-in">
          {seekIndicator}
        </div>
      )}

      {/* Контролы */}
        <div
        className={`absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/90 via-black/70 to-transparent backdrop-blur-sm transition-opacity duration-300 ${
            showControls || isDragging ? 'opacity-100' : 'opacity-0 pointer-events-none'
        }`}
        >
        <div 
            ref={progressRef}
            className="px-4 pt-3 pb-2 cursor-pointer"
            onMouseDown={handleProgressMouseDown}
        >
            <div className="w-full h-1.5 bg-white/20 rounded-full overflow-hidden hover:h-2 transition-all relative">
            <div
                className="absolute top-0 left-0 h-full bg-white/40 rounded-full transition-[width] duration-200"
                style={{ width: `${getBufferedPercent()}%` }}
            />
            <div
                className="absolute top-0 left-0 h-full bg-gradient-to-r from-blue-500 to-blue-600 rounded-full shadow-lg shadow-blue-500/50 transition-[width] duration-100"
                style={{ width: `${getProgressPercent()}%` }}
            >
                <div className="absolute right-0 top-1/2 transform translate-x-1/2 -translate-y-1/2 w-3 h-3 bg-white rounded-full shadow-lg" />
            </div>
            </div>
        </div>

        <div className="flex items-center gap-2 px-4 pb-3">
            <button
            onClick={togglePlay}
            className="flex items-center justify-center w-10 h-10 bg-white/10 hover:bg-white/20 rounded-lg backdrop-blur-sm transition-all"
            >
            {isPlaying ? <Pause size={20} className="text-white" /> : <Play size={20} className="text-white ml-0.5" />}
            </button>

            <div className="flex items-center gap-2">
            <button
                onClick={toggleMute}
                className="flex items-center justify-center w-10 h-10 bg-white/10 hover:bg-white/20 rounded-lg backdrop-blur-sm transition-all"
            >
                {isMuted || volume === 0 ? <VolumeX size={20} className="text-white" /> : <Volume2 size={20} className="text-white" />}
            </button>
            <input
                type="range"
                min="0"
                max="1"
                step="0.01"
                value={isMuted ? 0 : volume}
                onChange={handleVolumeChange}
                className="w-20 h-1.5 bg-white/20 rounded-full appearance-none cursor-pointer [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-3 [&::-webkit-slider-thumb]:h-3 [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-white [&::-webkit-slider-thumb]:cursor-pointer [&::-moz-range-thumb]:w-3 [&::-moz-range-thumb]:h-3 [&::-moz-range-thumb]:rounded-full [&::-moz-range-thumb]:bg-white [&::-moz-range-thumb]:border-0"
            />
            </div>

            {isLive && seekableEnd > 0 ? (
            <div className="flex items-center gap-2 px-3 py-1.5 bg-red-600 rounded-full ml-2">
                <span className="w-2 h-2 bg-white rounded-full animate-pulse" />
                <span className="text-white text-xs font-bold uppercase tracking-wide">LIVE</span>
            </div>
            ) : (
            <span className="text-white text-sm font-semibold ml-2">
                {formatTime(currentTime)} / {formatTime(seekableEnd)}
            </span>
            )}

            <div className="flex-1" />

            {/* ✅ Кнопка качества ЗДЕСЬ */}
            {qualities.length > 0 && (
            <div className="relative">
                <button
                onClick={() => setShowQualityMenu(!showQualityMenu)}
                className="flex items-center justify-center gap-1.5 w-auto h-10 px-3 bg-white/10 hover:bg-white/20 rounded-lg backdrop-blur-sm transition-all"
                >
                <Settings size={18} className="text-white" />
                <span className="text-white text-sm font-semibold">{currentQuality}</span>
                </button>

                {showQualityMenu && (
                <div className="absolute bottom-full right-0 mb-2 w-44 bg-gray-900/95 backdrop-blur-lg rounded-lg shadow-2xl overflow-hidden border border-gray-700">
                    <button
                    onClick={() => handleQualityChange('auto')}
                    className={`w-full px-4 py-3 text-left text-sm font-medium transition-colors ${
                        currentQuality === 'auto' ? 'bg-blue-600 text-white' : 'text-gray-200 hover:bg-gray-800'
                    }`}
                    >
                    Auto
                    </button>
                    {qualities.map((q) => (
                    <button
                        key={q.id}
                        onClick={() => handleQualityChange(q.id)}
                        className={`w-full px-4 py-3 text-left text-sm font-medium transition-colors ${
                        currentQuality === q.label ? 'bg-blue-600 text-white' : 'text-gray-200 hover:bg-gray-800'
                        }`}
                    >
                        {q.label}
                    </button>
                    ))}
                </div>
                )}
            </div>
            )}

            {/* Fullscreen */}
            <button
            onClick={toggleFullscreen}
            className="flex items-center justify-center w-10 h-10 bg-white/10 hover:bg-white/20 rounded-lg backdrop-blur-sm transition-all"
            >
            <Maximize size={20} className="text-white" />
            </button>
        </div>
    </div>

    </div>
  );
};
