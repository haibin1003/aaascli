// 在浏览器控制台执行此代码，导出能力平台数据
(function() {
    console.log('=== 数据导出工具 ===\n');
    
    // 1. 导出 Cookie 信息（不暴露完整值）
    console.log('【Cookie 状态】');
    const cookies = document.cookie.split(';');
    cookies.forEach(c => {
        const [name] = c.trim().split('=');
        console.log(`  ✓ ${name}: 已设置`);
    });
    
    // 2. 导出 localStorage
    console.log('\n【LocalStorage 数据】');
    const storageData = {};
    for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        let value = localStorage.getItem(key);
        // 尝试解析 JSON
        try {
            value = JSON.parse(value);
            console.log(`  ✓ ${key}: ${Array.isArray(value) ? 'Array(' + value.length + ')' : typeof value}`);
        } catch(e) {
            console.log(`  ✓ ${key}: string(${value.length})`);
        }
        storageData[key] = value;
    }
    
    // 3. 导出 Vuex Store 状态
    console.log('\n【Vuex Store 状态】');
    let storeData = {};
    try {
        const app = document.querySelector('#app').__vue__;
        if (app && app.$store) {
            storeData = app.$store.state;
            console.log('  Store Keys:', Object.keys(storeData));
        }
    } catch(e) {
        console.log('  无法访问 Store:', e.message);
    }
    
    // 4. 导出页面上的能力数据
    console.log('\n【页面能力数据】');
    const abilities = [];
    const cards = document.querySelectorAll('.el-card, .card, [class*="ability"]');
    cards.forEach((card, index) => {
        if (index >= 50) return; // 最多50个
        
        const title = card.querySelector('h1, h2, h3, h4, .title, [class*="title"]');
        const desc = card.querySelector('p, .desc, [class*="desc"]');
        
        // 尝试找"查看详情"链接
        const detailLink = card.querySelector('a[href*="detail"], button');
        let detailUrl = '';
        if (detailLink) {
            detailUrl = detailLink.getAttribute('href') || '';
        }
        
        if (title) {
            abilities.push({
                title: title.innerText.trim(),
                description: desc ? desc.innerText.trim().substring(0, 200) : '',
                detailUrl: detailUrl
            });
        }
    });
    console.log(`  找到 ${abilities.length} 个能力卡片`);
    
    // 5. 汇总结果
    const exportResult = {
        timestamp: new Date().toISOString(),
        url: window.location.href,
        pageTitle: document.title,
        cookies: cookies.map(c => c.split('=')[0].trim()), // 只返回名称
        localStorage: Object.keys(storageData),
        store: Object.keys(storeData),
        abilities: abilities.slice(0, 20), // 前20个
        fullStorage: storageData, // 完整数据
        fullStore: storeData // 完整数据
    };
    
    console.log('\n=== 导出完成 ===');
    console.log('请复制以下 JSON 数据给我（或右键→复制对象）：');
    console.log(JSON.stringify(exportResult, null, 2));
    
    // 自动复制
    navigator.clipboard.writeText(JSON.stringify(exportResult, null, 2)).then(() => {
        console.log('\n✅ 已复制到剪贴板！');
    }).catch(() => {
        console.log('\n⚠️ 自动复制失败，请手动复制上面的 JSON');
    });
    
    return exportResult;
})()
