// 存储应用程序状态
const state = {
    currentSong: null,
    playlist: [],
    favorites: [],
    searchResults: [],
    isPlaying: false,
    currentTime: 0,
    duration: 0,
    volume: 0.7,
};

// DOM 元素
const elements = {
    searchInput: document.getElementById('search-input'),
    searchButton: document.getElementById('search-button'),
    sourceCheckboxes: document.querySelectorAll('.source-checkbox'),
    menuItems: document.querySelectorAll('.menu-item'),
    tabContents: document.querySelectorAll('.tab-content'),
    searchResultsList: document.getElementById('search-results-list'),
    playlistSongs: document.getElementById('playlist-songs'),
    favoriteSongs: document.getElementById('favorite-songs'),
    resultsTotal: document.getElementById('results-total'),
    audioPlayer: document.getElementById('audio-player'),
    playButton: document.getElementById('play-button'),
    prevButton: document.getElementById('prev-button'),
    nextButton: document.getElementById('next-button'),
    currentSongCover: document.getElementById('current-song-cover'),
    currentSongTitle: document.getElementById('current-song-title'),
    currentSongArtist: document.getElementById('current-song-artist'),
    currentTime: document.getElementById('current-time'),
    totalTime: document.getElementById('total-time'),
    progressBar: document.querySelector('.progress-bar'),
    progress: document.querySelector('.progress'),
    volumeIcon: document.getElementById('volume-icon'),
    volumeSlider: document.querySelector('.volume-slider'),
    volumeProgress: document.querySelector('.volume-progress'),
};

// 初始化
function init() {
    // 设置音频播放器音量
    elements.audioPlayer.volume = state.volume;
    elements.volumeProgress.style.width = `${state.volume * 100}%`;
    
    // 从本地存储加载播放列表和收藏列表
    loadFromLocalStorage();
    
    // 渲染播放列表和收藏列表
    renderPlaylist();
    renderFavorites();
    
    // 添加事件监听器
    setupEventListeners();
}

// 从本地存储加载数据
function loadFromLocalStorage() {
    try {
        const playlist = localStorage.getItem('playlist');
        if (playlist) {
            state.playlist = JSON.parse(playlist);
        }
        
        const favorites = localStorage.getItem('favorites');
        if (favorites) {
            state.favorites = JSON.parse(favorites);
        }
    } catch (error) {
        console.error('Failed to load data from localStorage:', error);
    }
}

// 设置事件监听器
function setupEventListeners() {
    // 搜索按钮点击
    elements.searchButton.addEventListener('click', handleSearch);
    
    // 回车键搜索
    elements.searchInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            handleSearch();
        }
    });
    
    // 菜单项点击切换标签页
    elements.menuItems.forEach(item => {
        item.addEventListener('click', () => {
            const tabId = item.getAttribute('data-tab');
            switchTab(tabId, item);
        });
    });
    
    // 播放器控制
    elements.playButton.addEventListener('click', togglePlay);
    elements.prevButton.addEventListener('click', playPrevious);
    elements.nextButton.addEventListener('click', playNext);
    
    // 进度条控制
    elements.progressBar.addEventListener('click', seekAudio);
    
    // 音量控制
    elements.volumeSlider.addEventListener('click', changeVolume);
    elements.volumeIcon.addEventListener('click', toggleMute);
    
    // 音频播放事件
    elements.audioPlayer.addEventListener('timeupdate', updateProgress);
    elements.audioPlayer.addEventListener('ended', handleSongEnd);
    elements.audioPlayer.addEventListener('loadedmetadata', () => {
        state.duration = elements.audioPlayer.duration;
        elements.totalTime.textContent = formatTime(state.duration);
    });
}

// 处理搜索
async function handleSearch() {
    const searchTerm = elements.searchInput.value.trim();
    if (!searchTerm) return;
    
    // 获取选中的音乐源
    const selectedSources = [];
    elements.sourceCheckboxes.forEach(checkbox => {
        if (checkbox.checked) {
            selectedSources.push(checkbox.value);
        }
    });
    
    if (selectedSources.length === 0) {
        alert('请至少选择一个音乐源');
        return;
    }
    
    // 显示加载中指示器
    elements.searchResultsList.innerHTML = `
        <div class="loading">
            <div class="loading-spinner"></div>
        </div>
    `;
    
    try {
        // 发送搜索请求到后端API
        const response = await fetch(`/api/search?keyword=${encodeURIComponent(searchTerm)}&sources=${selectedSources.join(',')}`);
        
        if (!response.ok) {
            throw new Error('搜索请求失败');
        }
        
        const data = await response.json();
        state.searchResults = data.songs || [];
        
        // 更新搜索结果
        elements.resultsTotal.textContent = state.searchResults.length;
        renderSearchResults();
        
        // 切换到搜索结果标签页
        switchTab('search-results', document.querySelector('.menu-item[data-tab="search-results"]'));
    } catch (error) {
        console.error('搜索失败:', error);
        elements.searchResultsList.innerHTML = `
            <div class="empty-message">搜索失败，请稍后重试</div>
        `;
    }
}

// 渲染搜索结果
function renderSearchResults() {
    if (state.searchResults.length === 0) {
        elements.searchResultsList.innerHTML = `
            <div class="empty-message">未找到符合条件的音乐</div>
        `;
        return;
    }
    
    const html = state.searchResults.map((song, index) => createSongElement(song, index + 1, 'search')).join('');
    elements.searchResultsList.innerHTML = html;
    
    // 添加歌曲项事件
    addSongItemEvents('search');
}

// 渲染播放列表
function renderPlaylist() {
    if (state.playlist.length === 0) {
        elements.playlistSongs.innerHTML = `
            <div class="empty-message">播放列表为空</div>
        `;
        return;
    }
    
    const html = state.playlist.map((song, index) => createSongElement(song, index + 1, 'playlist')).join('');
    elements.playlistSongs.innerHTML = html;
    
    // 添加歌曲项事件
    addSongItemEvents('playlist');
}

// 渲染收藏列表
function renderFavorites() {
    if (state.favorites.length === 0) {
        elements.favoriteSongs.innerHTML = `
            <div class="empty-message">收藏列表为空</div>
        `;
        return;
    }
    
    const html = state.favorites.map((song, index) => createSongElement(song, index + 1, 'favorite')).join('');
    elements.favoriteSongs.innerHTML = html;
    
    // 添加歌曲项事件
    addSongItemEvents('favorite');
}

// 创建歌曲元素
function createSongElement(song, number, listType) {
    const isInPlaylist = state.playlist.some(item => item.id === song.id && item.source === song.source);
    const isInFavorites = state.favorites.some(item => item.id === song.id && item.source === song.source);
    
    const sourceName = getSourceName(song.source);
    const defaultCover = '/static/images/default-cover.jpg';
    
    return `
        <div class="song-item" data-id="${song.id}" data-source="${song.source}">
            <div class="song-number">${number}</div>
            <img class="song-cover" src="${song.cover || defaultCover}" onerror="this.src='${defaultCover}'" alt="${song.title}">
            <div class="song-info">
                <div class="song-title">${song.title}</div>
                <div class="song-artist">${song.artist}</div>
            </div>
            <div class="song-source">${sourceName}</div>
            <div class="song-actions">
                <button class="song-action-btn play-song">
                    <i class="fas fa-play"></i>
                </button>
                <button class="song-action-btn add-to-playlist" style="${isInPlaylist ? 'display:none;' : ''}">
                    <i class="fas fa-plus"></i>
                </button>
                <button class="song-action-btn remove-from-playlist" style="${listType === 'playlist' ? '' : 'display:none;'}">
                    <i class="fas fa-times"></i>
                </button>
                <button class="song-action-btn toggle-favorite">
                    <i class="fas ${isInFavorites ? 'fa-heart' : 'fa-heart-o'}"></i>
                </button>
            </div>
        </div>
    `;
}

// 添加歌曲项事件
function addSongItemEvents(listType) {
    const container = listType === 'search' ? elements.searchResultsList : 
                    listType === 'playlist' ? elements.playlistSongs :
                    elements.favoriteSongs;
    
    // 播放歌曲
    container.querySelectorAll('.play-song').forEach(button => {
        button.addEventListener('click', (e) => {
            const songItem = e.target.closest('.song-item');
            const songId = songItem.getAttribute('data-id');
            const songSource = songItem.getAttribute('data-source');
            
            const songList = listType === 'search' ? state.searchResults :
                           listType === 'playlist' ? state.playlist :
                           state.favorites;
            
            const song = songList.find(item => item.id === songId && item.source === songSource);
            if (song) {
                playSong(song);
            }
        });
    });
    
    // 添加到播放列表
    container.querySelectorAll('.add-to-playlist').forEach(button => {
        button.addEventListener('click', (e) => {
            const songItem = e.target.closest('.song-item');
            const songId = songItem.getAttribute('data-id');
            const songSource = songItem.getAttribute('data-source');
            
            const songList = listType === 'search' ? state.searchResults : state.favorites;
            const song = songList.find(item => item.id === songId && item.source === songSource);
            
            if (song) {
                addToPlaylist(song);
                button.style.display = 'none';
            }
        });
    });
    
    // 从播放列表中移除
    container.querySelectorAll('.remove-from-playlist').forEach(button => {
        button.addEventListener('click', (e) => {
            const songItem = e.target.closest('.song-item');
            const songId = songItem.getAttribute('data-id');
            const songSource = songItem.getAttribute('data-source');
            
            removeFromPlaylist(songId, songSource);
        });
    });
    
    // 收藏/取消收藏
    container.querySelectorAll('.toggle-favorite').forEach(button => {
        button.addEventListener('click', (e) => {
            const songItem = e.target.closest('.song-item');
            const songId = songItem.getAttribute('data-id');
            const songSource = songItem.getAttribute('data-source');
            
            const songList = listType === 'search' ? state.searchResults :
                           listType === 'playlist' ? state.playlist :
                           state.favorites;
            
            const song = songList.find(item => item.id === songId && item.source === songSource);
            
            if (song) {
                toggleFavorite(song, button);
            }
        });
    });
}

// 切换标签页
function switchTab(tabId, item) {
    // 更新激活的菜单项
    elements.menuItems.forEach(menuItem => {
        menuItem.classList.remove('active');
    });
    item.classList.add('active');
    
    // 更新显示的标签页
    elements.tabContents.forEach(content => {
        content.classList.remove('active');
    });
    document.getElementById(tabId).classList.add('active');
}

// 播放歌曲
async function playSong(song) {
    try {
        // 获取音乐URL
        const response = await fetch(`/api/song?id=${song.id}&source=${song.source}`);
        
        if (!response.ok) {
            throw new Error('获取音乐URL失败');
        }
        
        const data = await response.json();
        if (!data.url) {
            throw new Error('无法获取音乐URL');
        }
        
        // 更新当前播放歌曲信息
        state.currentSong = song;
        
        // 更新播放器UI
        elements.currentSongTitle.textContent = song.title;
        elements.currentSongArtist.textContent = song.artist;
        elements.currentSongCover.src = song.cover || '/static/images/default-cover.jpg';
        
        // 设置音频源并播放
        elements.audioPlayer.src = data.url;
        elements.audioPlayer.play()
            .then(() => {
                state.isPlaying = true;
                elements.playButton.innerHTML = '<i class="fas fa-pause"></i>';
            })
            .catch(error => {
                console.error('播放失败:', error);
                alert('播放失败，请尝试其他歌曲');
            });
    } catch (error) {
        console.error('播放歌曲失败:', error);
        alert('播放失败，请尝试其他歌曲');
    }
}

// 切换播放/暂停
function togglePlay() {
    if (!state.currentSong) return;
    
    if (state.isPlaying) {
        elements.audioPlayer.pause();
        state.isPlaying = false;
        elements.playButton.innerHTML = '<i class="fas fa-play"></i>';
    } else {
        elements.audioPlayer.play()
            .then(() => {
                state.isPlaying = true;
                elements.playButton.innerHTML = '<i class="fas fa-pause"></i>';
            })
            .catch(error => {
                console.error('播放失败:', error);
            });
    }
}

// 播放上一首
function playPrevious() {
    if (state.playlist.length === 0 || !state.currentSong) return;
    
    const currentIndex = state.playlist.findIndex(
        song => song.id === state.currentSong.id && song.source === state.currentSong.source
    );
    
    if (currentIndex === -1) {
        // 当前歌曲不在播放列表中，播放第一首
        playSong(state.playlist[0]);
    } else {
        // 播放上一首，如果是第一首则循环到最后一首
        const prevIndex = (currentIndex - 1 + state.playlist.length) % state.playlist.length;
        playSong(state.playlist[prevIndex]);
    }
}

// 播放下一首
function playNext() {
    if (state.playlist.length === 0 || !state.currentSong) return;
    
    const currentIndex = state.playlist.findIndex(
        song => song.id === state.currentSong.id && song.source === state.currentSong.source
    );
    
    if (currentIndex === -1) {
        // 当前歌曲不在播放列表中，播放第一首
        playSong(state.playlist[0]);
    } else {
        // 播放下一首，如果是最后一首则循环到第一首
        const nextIndex = (currentIndex + 1) % state.playlist.length;
        playSong(state.playlist[nextIndex]);
    }
}

// 更新进度条
function updateProgress() {
    state.currentTime = elements.audioPlayer.currentTime;
    const duration = elements.audioPlayer.duration || 0;
    
    // 更新进度条
    const progressPercent = (state.currentTime / duration) * 100;
    elements.progress.style.width = `${progressPercent}%`;
    
    // 更新时间显示
    elements.currentTime.textContent = formatTime(state.currentTime);
}

// 跳转到指定位置
function seekAudio(e) {
    const width = elements.progressBar.clientWidth;
    const clickX = e.offsetX;
    const duration = elements.audioPlayer.duration;
    
    elements.audioPlayer.currentTime = (clickX / width) * duration;
}

// 更改音量
function changeVolume(e) {
    const width = elements.volumeSlider.clientWidth;
    const clickX = e.offsetX;
    const volume = clickX / width;
    
    // 确保音量在0-1的范围内
    state.volume = Math.max(0, Math.min(1, volume));
    
    // 更新音频播放器音量
    elements.audioPlayer.volume = state.volume;
    
    // 更新音量条
    elements.volumeProgress.style.width = `${state.volume * 100}%`;
    
    // 更新音量图标
    updateVolumeIcon();
}

// 切换静音
function toggleMute() {
    if (elements.audioPlayer.volume > 0) {
        // 保存当前音量并设置为静音
        state.previousVolume = state.volume;
        state.volume = 0;
    } else {
        // 恢复之前的音量
        state.volume = state.previousVolume || 0.7;
    }
    
    // 更新音频播放器音量
    elements.audioPlayer.volume = state.volume;
    
    // 更新音量条
    elements.volumeProgress.style.width = `${state.volume * 100}%`;
    
    // 更新音量图标
    updateVolumeIcon();
}

// 更新音量图标
function updateVolumeIcon() {
    if (state.volume === 0) {
        elements.volumeIcon.className = 'fas fa-volume-mute';
    } else if (state.volume < 0.5) {
        elements.volumeIcon.className = 'fas fa-volume-down';
    } else {
        elements.volumeIcon.className = 'fas fa-volume-up';
    }
}

// 处理歌曲播放结束
function handleSongEnd() {
    playNext();
}

// 添加到播放列表
function addToPlaylist(song) {
    // 检查是否已经在播放列表中
    const exists = state.playlist.some(item => item.id === song.id && item.source === song.source);
    
    if (!exists) {
        state.playlist.push(song);
        
        // 保存到本地存储
        localStorage.setItem('playlist', JSON.stringify(state.playlist));
        
        // 更新播放列表UI
        renderPlaylist();
        
        // 如果当前没有播放歌曲，自动开始播放
        if (!state.currentSong) {
            playSong(song);
        }
    }
}

// 从播放列表中移除
function removeFromPlaylist(songId, songSource) {
    // 找到歌曲在播放列表中的索引
    const index = state.playlist.findIndex(song => song.id === songId && song.source === songSource);
    
    if (index !== -1) {
        // 判断是否正在播放此歌曲
        const isCurrentSong = state.currentSong && 
                            state.currentSong.id === songId && 
                            state.currentSong.source === songSource;
        
        // 从播放列表中移除
        state.playlist.splice(index, 1);
        
        // 保存到本地存储
        localStorage.setItem('playlist', JSON.stringify(state.playlist));
        
        // 更新播放列表UI
        renderPlaylist();
        
        // 如果正在播放此歌曲，播放下一首
        if (isCurrentSong && state.playlist.length > 0) {
            playNext();
        } else if (isCurrentSong) {
            // 停止播放
            elements.audioPlayer.pause();
            elements.audioPlayer.src = '';
            state.currentSong = null;
            state.isPlaying = false;
            elements.playButton.innerHTML = '<i class="fas fa-play"></i>';
            elements.currentSongTitle.textContent = '未播放';
            elements.currentSongArtist.textContent = '未知歌手';
            elements.currentSongCover.src = '/static/images/default-cover.jpg';
        }
    }
}

// 切换收藏状态
function toggleFavorite(song, button) {
    const iconElement = button.querySelector('i');
    
    // 检查歌曲是否已经在收藏列表中
    const existingIndex = state.favorites.findIndex(
        item => item.id === song.id && item.source === song.source
    );
    
    if (existingIndex === -1) {
        // 添加到收藏
        state.favorites.push(song);
        iconElement.className = 'fas fa-heart';
    } else {
        // 从收藏中移除
        state.favorites.splice(existingIndex, 1);
        iconElement.className = 'fas fa-heart-o';
        
        // 如果当前在收藏列表页面，需要更新UI
        if (document.getElementById('favorites').classList.contains('active')) {
            renderFavorites();
        }
    }
    
    // 保存到本地存储
    localStorage.setItem('favorites', JSON.stringify(state.favorites));
}

// 获取音乐来源名称
function getSourceName(source) {
    switch (source) {
        case 'qq':
            return '腾讯音乐';
        case 'netease':
            return '网易云';
        case 'kuwo':
            return '酷我';
        default:
            return source;
    }
}

// 格式化时间
function formatTime(seconds) {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = Math.floor(seconds % 60);
    return `${minutes}:${remainingSeconds < 10 ? '0' : ''}${remainingSeconds}`;
}

// 初始化应用
document.addEventListener('DOMContentLoaded', init); 