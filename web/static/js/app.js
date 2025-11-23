// TrackMyTime Dashboard with Tailwind

const API_BASE = 'http://localhost:8787';
const REFRESH_INTERVAL = 5000;

let currentPeriod = 'today';
let refreshTimer = null;
let countdownTimer = null;
let donutChart = null;
let timelineChart = null;
let isGroupedView = true; // Vue groupée par défaut
let openGroups = new Set(); // Garder trace des groupes ouverts

// ============================================
// Initialization
// ============================================

document.addEventListener('DOMContentLoaded', () => {
    console.log('⏱️ Dashboard TrackMyTime initialisé');
    initCharts();
    refreshDashboard();
    startAutoRefresh();
    checkAPIHealth();
});

// ============================================
// API Functions
// ============================================

async function fetchAPI(endpoint) {
    try {
        const response = await fetch(`${API_BASE}${endpoint}`);
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        return await response.json();
    } catch (error) {
        console.error(`Erreur API (${endpoint}):`, error);
        updateStatus('offline', 'Hors ligne');
        throw error;
    }
}

async function checkAPIHealth() {
    try {
        await fetchAPI('/health');
        updateStatus('online', 'En ligne');
    } catch (error) {
        updateStatus('offline', 'Hors ligne');
    }
}

// ============================================
// Dashboard Refresh
// ============================================

async function refreshDashboard() {
    console.log(`🔄 Refreshing (${currentPeriod})`);
    
    const refreshIcon = document.getElementById('refresh-icon');
    refreshIcon.style.animation = 'spin 1s linear infinite';
    
    try {
        await Promise.all([
            updateCurrentActivity(),
            updateStats(),
            updateTopApps()
        ]);
        updateStatus('online', 'En ligne');
    } catch (error) {
        console.error('Échec du rafraîchissement:', error);
        updateStatus('offline', 'Erreur');
    } finally {
        refreshIcon.style.animation = '';
    }
}

async function updateCurrentActivity() {
    try {
        const data = await fetchAPI('/activity/current');
        const currentApp = document.getElementById('current-app');
        
        if (data.status === 'no activity') {
            currentApp.textContent = '--';
        } else {
            currentApp.textContent = data.app_name;
            currentApp.title = data.window_title;
        }
    } catch (error) {
        console.error('Échec de la mise à jour de l\'activité actuelle:', error);
    }
}

async function updateStats() {
    try {
        const endpoint = currentPeriod === 'today' ? '/stats/today' : '/stats/week';
        const data = await fetchAPI(endpoint);
        
        document.getElementById('total-time').textContent = formatDuration(data.total_active_seconds);
        document.getElementById('idle-time').textContent = formatDuration(data.total_idle_seconds || 0);
        document.getElementById('apps-count').textContent = Object.keys(data.stats_by_app).length;
        document.getElementById('period-label').textContent = currentPeriod === 'today' ? 'Aujourd\'hui' : 'Cette semaine';
        
        updateDonutChart(data.stats_by_app);
        await updateTimelineChart();
        
    } catch (error) {
        console.error('Échec de la mise à jour des stats:', error);
    }
}

async function updateTopApps() {
    try {
        if (isGroupedView) {
            const endpoint = `/api/stats/grouped?period=${currentPeriod}`;
            const data = await fetchAPI(endpoint);
            renderGroupedApps(data.groups);
        } else {
            const endpoint = currentPeriod === 'today' ? '/stats/today' : '/stats/week';
            const data = await fetchAPI(endpoint);
            renderFlatApps(data.stats_by_app);
        }
    } catch (error) {
        console.error('Échec de la mise à jour des top apps:', error);
    }
}

function toggleGroupView() {
    isGroupedView = !isGroupedView;
    const btn = document.getElementById('toggle-group-btn');
    btn.textContent = isGroupedView ? '📋 Vue simple' : '👁️ Vue groupée';
    updateTopApps();
}

function renderGroupedApps(groups) {
    const list = document.getElementById('apps-list');
    
    if (!groups || groups.length === 0) {
        list.innerHTML = '<div class="text-center py-12 text-gray-400">Aucune donnée disponible</div>';
        return;
    }
    
    list.innerHTML = groups.map((group, groupIndex) => {
        const totalDuration = formatDuration(group.total_seconds);
        const hasChildren = group.children.length > 1 || 
                           (group.children.length === 1 && group.children[0].name !== group.app_name);
        
        let rankBg = 'bg-gray-100 text-gray-600';
        if (groupIndex === 0) rankBg = 'bg-gradient-to-br from-yellow-400 to-orange-500 text-white';
        else if (groupIndex === 1) rankBg = 'bg-gradient-to-br from-gray-300 to-gray-400 text-gray-800';
        else if (groupIndex === 2) rankBg = 'bg-gradient-to-br from-orange-400 to-red-500 text-white';
        
        // Vérifier si ce groupe était ouvert avant le refresh
        const isOpen = openGroups.has(groupIndex);
        
        const childrenHtml = hasChildren ? `
            <div class="ml-16 mt-2 space-y-2 border-l-2 border-gray-200 pl-4" id="children-${groupIndex}" style="display: ${isOpen ? 'block' : 'none'};">
                ${group.children.map(child => {
                    const childDuration = formatDuration(child.duration);
                    const percentage = ((child.duration / group.total_seconds) * 100).toFixed(1);
                    return `
                        <div class="flex items-center justify-between p-2 rounded hover:bg-gray-50 transition-all">
                            <div class="flex-1">
                                <span class="text-sm font-medium text-gray-700">${escapeHtml(child.name)}</span>
                                <div class="w-full h-1.5 bg-gray-100 rounded-full mt-1 overflow-hidden">
                                    <div class="h-full bg-indigo-500 rounded-full transition-all duration-700" style="width: ${percentage}%"></div>
                                </div>
                            </div>
                            <div class="ml-4 text-right shrink-0">
                                <span class="text-sm font-mono font-semibold text-gray-600">${childDuration}</span>
                                <span class="text-xs text-gray-400 ml-2">${percentage}%</span>
                            </div>
                        </div>
                    `;
                }).join('')}
            </div>
        ` : '';
        
        return `
            <div class="border border-gray-100 rounded-lg overflow-hidden hover:border-gray-200 hover:shadow-md transition-all">
                <div class="flex items-center gap-4 p-4 ${hasChildren ? 'cursor-pointer' : ''}" 
                     ${hasChildren ? `onclick="toggleChildren(${groupIndex})"` : ''}>
                    <div class="w-12 h-12 ${rankBg} rounded-lg flex items-center justify-center font-bold text-lg shrink-0">
                        ${groupIndex + 1}
                    </div>
                    
                    <div class="flex-1 min-w-0">
                        <p class="font-semibold text-gray-900 truncate">${escapeHtml(group.app_name)}</p>
                        <p class="text-sm text-gray-500">${group.children.length} activité${group.children.length > 1 ? 's' : ''}</p>
                    </div>
                    
                    <div class="text-right shrink-0">
                        <p class="font-mono font-semibold text-gray-900">${totalDuration}</p>
                    </div>
                    
                    ${hasChildren ? `
                        <svg class="w-5 h-5 text-gray-400 transition-transform" id="icon-${groupIndex}" style="transform: rotate(${isOpen ? '180' : '0'}deg);" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
                        </svg>
                    ` : ''}
                </div>
                ${childrenHtml}
            </div>
        `;
    }).join('');
}

function toggleChildren(index) {
    const childrenDiv = document.getElementById(`children-${index}`);
    const icon = document.getElementById(`icon-${index}`);
    
    if (childrenDiv.style.display === 'none') {
        childrenDiv.style.display = 'block';
        icon.style.transform = 'rotate(180deg)';
        openGroups.add(index); // Sauvegarder l'état ouvert
    } else {
        childrenDiv.style.display = 'none';
        icon.style.transform = 'rotate(0deg)';
        openGroups.delete(index); // Retirer de l'état ouvert
    }
}

function renderFlatApps(stats) {
    const sortedApps = Object.entries(stats)
        .sort((a, b) => b[1] - a[1]);
    
    const totalSeconds = Object.values(stats).reduce((sum, val) => sum + val, 0);
    const list = document.getElementById('apps-list');
    
    if (sortedApps.length === 0) {
        list.innerHTML = '<div class="text-center py-12 text-gray-400">Aucune donnée disponible</div>';
        return;
    }
    
    list.innerHTML = sortedApps.map(([app, seconds], index) => {
        const percentage = totalSeconds > 0 ? (seconds / totalSeconds * 100).toFixed(1) : 0;
        const duration = formatDuration(seconds);
        const totalDuration = formatDuration(seconds);
        
        let rankBg = 'bg-gray-100 text-gray-600';
        if (index === 0) rankBg = 'bg-gradient-to-br from-yellow-400 to-orange-500 text-white';
        else if (index === 1) rankBg = 'bg-gradient-to-br from-gray-300 to-gray-400 text-gray-800';
        else if (index === 2) rankBg = 'bg-gradient-to-br from-orange-400 to-red-500 text-white';
        
        return `
            <div class="flex items-center gap-4 p-4 rounded-lg border border-gray-100 hover:border-gray-200 hover:shadow-md transition-all group">
                <div class="w-12 h-12 ${rankBg} rounded-lg flex items-center justify-center font-bold text-lg shrink-0">
                    ${index + 1}
                </div>
                
                <div class="flex-1 min-w-0">
                    <p class="font-semibold text-gray-900 truncate">${escapeHtml(app)}</p>
                    <p class="text-sm text-gray-500">${totalDuration} total</p>
                </div>
                
                <div class="text-right shrink-0">
                    <p class="font-mono font-semibold text-gray-900">${duration}</p>
                    <div class="flex items-center gap-2 mt-1">
                        <div class="w-24 h-2 bg-gray-100 rounded-full overflow-hidden">
                            <div class="h-full bg-gradient-to-r from-indigo-500 to-purple-600 rounded-full transition-all duration-700" style="width: ${percentage}%"></div>
                        </div>
                        <span class="text-sm text-gray-500 font-medium">${percentage}%</span>
                    </div>
                </div>
            </div>
        `;
    }).join('');
}

// ============================================
// Chart Functions
// ============================================

function initCharts() {
    // Timeline Chart
    const timelineCtx = document.getElementById('timeline-chart').getContext('2d');
    timelineChart = new Chart(timelineCtx, {
        type: 'bar',
        data: {
            labels: Array.from({length: 24}, (_, i) => `${String(i).padStart(2, '0')}:00`),
            datasets: [{
                label: 'Temps Actif (minutes)',
                data: new Array(24).fill(0),
                backgroundColor: 'rgba(99, 102, 241, 0.8)',
                borderColor: 'rgb(99, 102, 241)',
                borderWidth: 0,
                borderRadius: 6,
                barThickness: 20
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: { display: false },
                tooltip: {
                    backgroundColor: 'rgba(17, 24, 39, 0.95)',
                    padding: 12,
                    titleColor: '#fff',
                    bodyColor: '#fff',
                    borderColor: 'rgba(99, 102, 241, 0.5)',
                    borderWidth: 1,
                    displayColors: false,
                    callbacks: {
                        label: (context) => {
                            const minutes = context.parsed.y;
                            return `Actif: ${minutes} min (${(minutes / 60).toFixed(1)}h)`;
                        }
                    }
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    grid: { color: 'rgba(0, 0, 0, 0.05)' },
                    ticks: { color: '#6b7280' }
                },
                x: {
                    grid: { display: false },
                    ticks: { color: '#6b7280' }
                }
            }
        }
    });
    
    // Donut Chart
    const donutCtx = document.getElementById('donut-chart').getContext('2d');
    donutChart = new Chart(donutCtx, {
        type: 'doughnut',
        data: {
            labels: [],
            datasets: [{
                data: [],
                backgroundColor: [
                    '#6366f1', '#8b5cf6', '#ec4899', '#f43f5e',
                    '#f59e0b', '#10b981', '#06b6d4', '#3b82f6'
                ],
                borderWidth: 0,
                spacing: 2
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    position: 'bottom',
                    labels: {
                        padding: 15,
                        usePointStyle: true,
                        pointStyle: 'circle',
                        font: { size: 12, weight: '500' },
                        color: '#374151'
                    }
                },
                tooltip: {
                    backgroundColor: 'rgba(17, 24, 39, 0.95)',
                    padding: 12,
                    titleColor: '#fff',
                    bodyColor: '#fff',
                    borderColor: 'rgba(99, 102, 241, 0.5)',
                    borderWidth: 1,
                    callbacks: {
                        label: (context) => {
                            const label = context.label || '';
                            const value = context.parsed || 0;
                            const duration = formatDuration(value);
                            const total = context.dataset.data.reduce((a, b) => a + b, 0);
                            const percent = ((value / total) * 100).toFixed(1);
                            return `${label}: ${duration} (${percent}%)`;
                        }
                    }
                }
            }
        }
    });
}

function updateDonutChart(stats) {
    const sortedApps = Object.entries(stats)
        .sort((a, b) => b[1] - a[1])
        .slice(0, 8);
    
    const labels = sortedApps.map(([app]) => app);
    const data = sortedApps.map(([, seconds]) => seconds);
    
    donutChart.data.labels = labels;
    donutChart.data.datasets[0].data = data;
    donutChart.update();
}

async function updateTimelineChart() {
    try {
        const data = await fetchAPI('/api/stats/hourly?period=' + currentPeriod);
        const minutes = data.hourly_data.map(seconds => Math.round(seconds / 60));
        
        timelineChart.data.datasets[0].data = minutes;
        timelineChart.update();
    } catch (error) {
        console.error('Échec de la mise à jour de la timeline:', error);
    }
}

// ============================================
// Export & Controls
// ============================================

function exportData(format) {
    const url = `${API_BASE}/export/aggregated?period=${currentPeriod}&format=${format}`;
    const link = document.createElement('a');
    link.href = url;
    link.download = `trackmytime_${currentPeriod}_${getDateString()}.${format}`;
    link.click();
}

function setPeriod(period) {
    currentPeriod = period;
    
    const btnToday = document.getElementById('btn-today');
    const btnWeek = document.getElementById('btn-week');
    
    if (period === 'today') {
        btnToday.className = 'px-4 py-2 rounded-md text-sm font-medium transition-all bg-gray-900 text-white';
        btnWeek.className = 'px-4 py-2 rounded-md text-sm font-medium text-gray-600 hover:bg-gray-100 transition-all';
    } else {
        btnWeek.className = 'px-4 py-2 rounded-md text-sm font-medium transition-all bg-gray-900 text-white';
        btnToday.className = 'px-4 py-2 rounded-md text-sm font-medium text-gray-600 hover:bg-gray-100 transition-all';
    }
    
    refreshDashboard();
}

// ============================================
// Auto Refresh
// ============================================

function startAutoRefresh() {
    let countdown = 5;
    
    countdownTimer = setInterval(() => {
        countdown--;
        document.getElementById('countdown').textContent = countdown;
        if (countdown <= 0) countdown = 5;
    }, 1000);
    
    refreshTimer = setInterval(() => {
        refreshDashboard();
    }, REFRESH_INTERVAL);
}

// ============================================
// Utilities
// ============================================

function updateStatus(status, text) {
    const dot = document.getElementById('status-dot');
    const textEl = document.getElementById('status-text');
    
    if (status === 'online') {
        dot.className = 'w-2 h-2 rounded-full bg-green-500 animate-pulse';
    } else {
        dot.className = 'w-2 h-2 rounded-full bg-gray-300';
    }
    
    textEl.textContent = text;
}

function formatDuration(seconds) {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    return `${pad(hours)}:${pad(minutes)}:${pad(secs)}`;
}

function pad(num) {
    return String(num).padStart(2, '0');
}

function escapeHtml(text) {
    const map = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#039;' };
    return text.replace(/[&<>"']/g, m => map[m]);
}

function getDateString() {
    const now = new Date();
    return `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}`;
}

// Cleanup
window.addEventListener('beforeunload', () => {
    if (refreshTimer) clearInterval(refreshTimer);
    if (countdownTimer) clearInterval(countdownTimer);
});

console.log('✅ Dashboard prêt');
