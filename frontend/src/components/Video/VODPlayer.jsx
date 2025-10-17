import { useEffect, useRef, useState } from 'react';
import videojs from 'video.js';
import 'video.js/dist/video-js.css';
import { Play, Pause, Volume2, VolumeX, Maximize, Settings } from 'lucide-react';

export const VODPlayer = ({ videoUrl, autoplay = false, startTime = 0 }) => {
  const videoRef = useRef(null);
  const playerRef = useRef(null);
  const containerRef = useRef(null);
  const progressRef = useRef(null);
  
  const [isPlaying, setIsPlaying] = useState(false);
  const [volume, setVolume] = useState(1);
  const [isMuted, setIsMuted] = useState(false);
  const [previousVolume, setPreviousVolume] = useState(1);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [bufferedEnd, setBufferedEnd] = useState(0);
  const [showControls, setShowControls] = useState(true);
  const [qualities, setQualities] = useState([]);
  const [currentQuality, setCurrentQuality] = useState('auto');
  const [showQualityMenu, setShowQualityMenu] = useState(false);
  const [isDragging, setIsDragging] = useState(false);
  const [seekIndicator, setSeekIndicator] = useState(null);
  
  let hideControlsTimeout = useRef(null);

  useEffect(() => {
    const initPlayer = () => {
      if (!videoRef.current || playerRef.current) return;

      console.log('🎥 Initializing VOD player with URL:', videoUrl);

      const player = videojs(videoRef.current, {
        controls: false,
        autoplay: autoplay,
        muted: false,
        preload: 'auto',
        fluid: true,
        html5: {
          vhs: {
            overrideNative: true,
          },
        },
      });

      playerRef.current = player;

      const isHLS = videoUrl.endsWith('.m3u8') || videoUrl.includes('master.m3u8');
      
      console.log('📺 Loading video:', videoUrl);
      console.log('📺 Type:', isHLS ? 'HLS' : 'MP4');
      console.log('🍪 Auth: via cookie (automatic)');
      
      player.src({
        src: videoUrl,
        type: isHLS ? 'application/x-mpegURL' : 'video/mp4',
      });

      player.ready(() => {
        console.log('✅ VOD Player ready');

        if (isHLS) {
          const qualityLevels = player.qualityLevels();
          if (qualityLevels) {
            qualityLevels.on('addqualitylevel', () => updateQualityList(qualityLevels));
            updateQualityList(qualityLevels);
          }
        }

        if (autoplay) {
          const playPromise = player.play();
          if (playPromise !== undefined) {
            playPromise.catch(error => {
              console.log('⚠️ Autoplay blocked, trying muted:', error);
              player.muted(true);
              player.play().catch(e => console.error('❌ Failed to play:', e));
            });
          }
        }
      });

      player.on('play', () => setIsPlaying(true));
      player.on('pause', () => setIsPlaying(false));
      
      player.on('volumechange', () => {
        const currentVol = player.volume();
        const muted = player.muted();
        setVolume(currentVol);
        setIsMuted(muted);
        if (!muted && currentVol > 0) {
          setPreviousVolume(currentVol);
        }
      });

      player.on('loadedmetadata', () => {
        const dur = player.duration();
        setDuration(dur);
        console.log('📊 Video duration:', dur);
        if (startTime > 0) {
          player.currentTime(startTime);
        }
      });

      player.on('timeupdate', () => {
        if (!isDragging) {
          setCurrentTime(player.currentTime());
        }
        const buffered = player.buffered();
        if (buffered.length > 0) {
          setBufferedEnd(buffered.end(buffered.length - 1));
        }
      });

      player.on('error', () => {
        const error = player.error();
        console.error('❌ Player error:', error);
        console.error('Error code:', error?.code);
        console.error('Error message:', error?.message);
      });

      player.on('loadeddata', () => {
        console.log('✅ Video loaded successfully');
      });
    };

    const timer = setTimeout(initPlayer, 0);

    return () => {
      clearTimeout(timer);
      if (playerRef.current && !playerRef.current.isDisposed()) {
        playerRef.current.dispose();
        playerRef.current = null;
      }
    };
  }, [videoUrl, autoplay, startTime, isDragging]);

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

  useEffect(() => {
    const handleKeyDown = (e) => {
      if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return;
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
          seek(-5);
          break;

        case 'ArrowRight':
          e.preventDefault();
          seek(5);
          break;

        case 'j':
        case 'J':
          e.preventDefault();
          seek(-10);
          break;

        case 'l':
        case 'L':
          e.preventDefault();
          seek(10);
          break;

        case 'ArrowUp':
          e.preventDefault();
          changeVolume(0.1);
          break;

        case 'ArrowDown':
          e.preventDefault();
          changeVolume(-0.1);
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
          const newTime = duration * percent;
          playerRef.current.currentTime(newTime);
          setCurrentTime(newTime);
          break;

        case 'Home':
          e.preventDefault();
          playerRef.current.currentTime(0);
          setCurrentTime(0);
          break;

        case 'End':
          e.preventDefault();
          playerRef.current.currentTime(duration - 1);
          setCurrentTime(duration - 1);
          break;

        default:
          break;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [currentTime, duration, volume, isMuted]);

  const togglePlay = () => {
    if (!playerRef.current) return;
    if (isPlaying) {
      playerRef.current.pause();
    } else {
      playerRef.current.play();
    }
  };

  const seek = (seconds) => {
    if (!playerRef.current) return;
    const newTime = Math.max(0, Math.min(duration, currentTime + seconds));
    playerRef.current.currentTime(newTime);
    setCurrentTime(newTime);
    setSeekIndicator(seconds > 0 ? `+${seconds}s` : `${seconds}s`);
    setTimeout(() => setSeekIndicator(null), 500);
  };

  const changeVolume = (delta) => {
    if (!playerRef.current) return;
    const newVolume = Math.max(0, Math.min(1, volume + delta));
    playerRef.current.volume(newVolume);
    if (isMuted && delta > 0) {
      playerRef.current.muted(false);
    }
  };

  const toggleMute = () => {
    if (!playerRef.current) return;
    playerRef.current.muted(!isMuted);
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
    if (!playerRef.current || !progressRef.current || duration === 0) return;
    const rect = progressRef.current.getBoundingClientRect();
    const pos = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width));
    const newTime = pos * duration;
    playerRef.current.currentTime(newTime);
    setCurrentTime(newTime);
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
  }, [isDragging, duration]);

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
    if (duration <= 0) return 0;
    return (currentTime / duration) * 100;
  };

  const getBufferedPercent = () => {
    if (duration <= 0) return 0;
    return (bufferedEnd / duration) * 100;
  };

  return (
    <div
      ref={containerRef}
      className="relative w-full bg-black rounded-lg overflow-hidden shadow-2xl group"
      onMouseMove={handleMouseMove}
      onMouseLeave={() => isPlaying && !isDragging && setShowControls(false)}
    >
      <div data-vjs-player onClick={togglePlay} className="cursor-pointer">
        <video ref={videoRef} className="video-js w-full" />
      </div>

      {seekIndicator && (
        <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 bg-black/80 text-white px-6 py-3 rounded-lg text-2xl font-bold backdrop-blur-sm pointer-events-none z-40">
          {seekIndicator}
        </div>
      )}

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
              onChange={(e) => {
                const newVolume = parseFloat(e.target.value);
                if (playerRef.current) {
                  playerRef.current.volume(newVolume);
                  if (newVolume > 0 && isMuted) {
                    playerRef.current.muted(false);
                  }
                }
              }}
              className="w-20 h-1.5 bg-white/20 rounded-full appearance-none cursor-pointer [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-3 [&::-webkit-slider-thumb]:h-3 [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-white [&::-webkit-slider-thumb]:cursor-pointer [&::-moz-range-thumb]:w-3 [&::-moz-range-thumb]:h-3 [&::-moz-range-thumb]:rounded-full [&::-moz-range-thumb]:bg-white [&::-moz-range-thumb]:border-0"
            />
          </div>

          <span className="text-white text-sm font-semibold ml-2">
            {formatTime(currentTime)} / {formatTime(duration)}
          </span>

          <div className="flex-1" />

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
