// ============================================
// 脚本 1：导出数字服务页面完整数据
// 在 https://service.sd.10086.cn/aaas/#/sdOpenPortal/digitalServices 执行
// ============================================

(function() {
    console.log('=== 数字服务页面数据导出 ===\n');
    
    var result = {
        exportTime: new Date().toISOString(),
        pageUrl: window.location.href,
        pageTitle: document.title
    };
    
    // 1. 标签页信息（对内/对外/网络域）
    console.log('【1. 服务类型标签】');
    result.serviceTypes = [];
    document.querySelectorAll('li, .tab, [role="tab"]').forEach(function(el) {
        var text = el.innerText || el.textContent || '';
        if (text.includes('对内') || text.includes('对外') || text.includes('网络')) {
            result.serviceTypes.push({
                name: text.trim(),
                className: el.className,
                isActive: el.classList.contains('active') || el.className.includes('active')
            });
        }
    });
    console.log('找到', result.serviceTypes.length, '个标签:', result.serviceTypes.map(function(t) { return t.name; }));
    
    // 2. 分类/目录结构
    console.log('\n【2. 服务分类/目录】');
    result.categories = [];
    document.querySelectorAll('.category, .catalog, .level1, .level2, .el-submenu__title, .menu-item').forEach(function(el) {
        var text = el.innerText || el.textContent || '';
        if (text.trim() && text.trim().length < 50 && !text.includes('服务')) {
            result.categories.push({
                name: text.trim(),
                className: el.className
            });
        }
    });
    // 去重
    result.categories = result.categories.filter(function(item, index, self) {
        return index === self.findIndex(function(t) { return t.name === item.name; });
    });
    console.log('找到', result.categories.length, '个分类');
    result.categories.slice(0, 10).forEach(function(c) { console.log('  -', c.name); });
    
    // 3. 服务列表（从 DOM 提取）
    console.log('\n【3. 服务列表（前50个）】');
    result.services = [];
    var seenServices = new Set();
    
    // 查找所有可能的服务项
    document.querySelectorAll('*').forEach(function(el) {
        var text = el.innerText || el.textContent || '';
        
        // 服务名称特征：纯文本、无子元素、长度适中、不包含特定关键词
        if (el.children.length === 0 && 
            text.trim().length > 5 && 
            text.trim().length < 100 &&
            !text.includes('\n') &&
            !text.includes('服务') &&
            !text.includes('目录') &&
            !text.includes('筛选') &&
            !text.includes('Copyright') &&
            !text.includes('加载中') &&
            !seenServices.has(text.trim())) {
            
            seenServices.add(text.trim());
            result.services.push({
                name: text.trim(),
                element: el.tagName
            });
        }
    });
    
    console.log('找到', result.services.length, '个服务');
    result.services.slice(0, 10).forEach(function(s) { console.log('  -', s.name); });
    
    // 4. 筛选器/下拉菜单
    console.log('\n【4. 筛选器】');
    result.filters = [];
    document.querySelectorAll('.filter, .el-select, .dropdown, [class*="select"]').forEach(function(el) {
        var text = el.innerText || el.textContent || '';
        if (text.includes('业务域') || text.includes('目录') || text.includes('筛选')) {
            result.filters.push({
                name: text.trim().substring(0, 50),
                className: el.className
            });
        }
    });
    console.log('找到', result.filters.length, '个筛选器');
    
    // 5. Vue 数据（深层）
    console.log('\n【5. Vue 实例数据】');
    result.vueData = [];
    var allElements = document.querySelectorAll('*');
    
    for (var i = 0; i < allElements.length && result.vueData.length < 3; i++) {
        var vue = allElements[i].__vue__;
        if (vue && vue.$data) {
            var keys = Object.keys(vue.$data);
            var arrays = keys.filter(function(k) {
                return Array.isArray(vue.$data[k]) && vue.$data[k].length > 20;
            });
            
            if (arrays.length > 0) {
                var dataInfo = {
                    elementIndex: i,
                    arrayFields: []
                };
                
                arrays.forEach(function(k) {
                    var arr = vue.$data[k];
                    dataInfo.arrayFields.push({
                        key: k,
                        length: arr.length,
                        sample: arr.length > 0 ? {
                            id: arr[0].id || arr[0].serviceId || arr[0].code || arr[0].apiId,
                            name: arr[0].name || arr[0].serviceName || arr[0].apiName || arr[0].title,
                            type: arr[0].type || arr[0].serviceType,
                            allKeys: Object.keys(arr[0]).slice(0, 10)
                        } : null
                    });
                });
                
                result.vueData.push(dataInfo);
            }
        }
    }
    console.log('找到', result.vueData.length, '个 Vue 数据对象');
    
    // 6. 按钮和操作
    console.log('\n【6. 操作按钮】');
    result.buttons = [];
    document.querySelectorAll('button, a, .btn, [class*="button"]').forEach(function(el) {
        var text = el.innerText || el.textContent || '';
        if (text.includes('订购') || text.includes('授权') || text.includes('申请') || 
            text.includes('详情') || text.includes('查看') || text.includes('使用') ||
            text.includes('订阅') || text.includes('开通')) {
            result.buttons.push({
                text: text.trim(),
                tag: el.tagName,
                className: el.className,
                href: el.getAttribute('href') || ''
            });
        }
    });
    console.log('找到', result.buttons.length, '个操作按钮');
    result.buttons.forEach(function(b) { console.log('  -', b.text, '(' + b.tag + ')'); });
    
    // 7. 导出完整结果
    console.log('\n=== 导出完成 ===');
    console.log(JSON.stringify(result, null, 2));
    
    // 复制到剪贴板
    navigator.clipboard.writeText(JSON.stringify(result, null, 2)).then(function() {
        console.log('\n✅ 已复制到剪贴板！请粘贴给我');
    }).catch(function() {
        console.log('\n⚠️ 自动复制失败，请手动复制上面的 JSON');
    });
    
    return result;
})();
