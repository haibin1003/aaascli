// 在数字服务页面执行此代码
(function() {
    console.log('=== 导出数字服务数据 ===\n');
    
    var result = {
        timestamp: new Date().toISOString(),
        url: window.location.href,
        pageTitle: document.title
    };
    
    // 1. 查找所有 Vue 实例中的数据
    console.log('【1. 查找页面数据】');
    var allElements = document.querySelectorAll('*');
    var pageData = [];
    
    for (var i = 0; i < allElements.length; i++) {
        var vue = allElements[i].__vue__;
        if (vue && vue.$data) {
            var keys = Object.keys(vue.$data);
            var dataKeys = keys.filter(function(k) {
                var lower = k.toLowerCase();
                return lower.includes('service') || 
                       lower.includes('api') || 
                       lower.includes('list') || 
                       lower.includes('data') ||
                       lower.includes('catalog') ||
                       lower.includes('category');
            });
            
            if (dataKeys.length > 0) {
                var dataInfo = {
                    elementIndex: i,
                    keys: keys,
                    dataKeys: dataKeys,
                    dataSample: {}
                };
                
                dataKeys.forEach(function(k) {
                    var val = vue.$data[k];
                    if (Array.isArray(val)) {
                        dataInfo.dataSample[k] = {
                            type: 'array',
                            length: val.length,
                            firstItem: val.length > 0 ? JSON.stringify(val[0]).substring(0, 300) : null
                        };
                    } else if (typeof val === 'object' && val !== null) {
                        dataInfo.dataSample[k] = {
                            type: 'object',
                            keys: Object.keys(val)
                        };
                    }
                });
                
                pageData.push(dataInfo);
            }
        }
    }
    
    result.pageData = pageData;
    console.log('找到 ' + pageData.length + ' 个数据对象');
    
    // 2. 区分对内/对外服务
    console.log('\n【2. 服务类型分析】');
    var serviceTypes = {
        internal: [], // 对内服务
        external: []  // 对外服务
    };
    
    // 查找页面上的标签或筛选器
    document.querySelectorAll('.tab, .filter, .tag, button').forEach(function(el) {
        var text = el.innerText || el.textContent || '';
        if (text.includes('对内') || text.includes('内部')) {
            serviceTypes.internal.push(text.trim());
        }
        if (text.includes('对外') || text.includes('外部') || text.includes('开放')) {
            serviceTypes.external.push(text.trim());
        }
    });
    
    result.serviceTypes = serviceTypes;
    console.log('对内服务标签:', serviceTypes.internal);
    console.log('对外服务标签:', serviceTypes.external);
    
    // 3. 查找服务列表
    console.log('\n【3. 服务列表】');
    var services = [];
    
    // 尝试从 Vue 数据中提取
    pageData.forEach(function(pd) {
        Object.keys(pd.dataSample).forEach(function(key) {
            var sample = pd.dataSample[key];
            if (sample.type === 'array' && sample.length > 0 && sample.firstItem) {
                // 可能是服务列表
                services.push({
                    dataKey: key,
                    count: sample.length,
                    sample: sample.firstItem
                });
            }
        });
    });
    
    result.services = services;
    console.log('找到 ' + services.length + ' 个服务列表');
    
    // 4. 查找订购/授权相关的按钮或链接
    console.log('\n【4. 操作按钮】');
    var actions = [];
    document.querySelectorAll('button, a').forEach(function(el) {
        var text = el.innerText || el.textContent || '';
        if (text.includes('订购') || text.includes('授权') || text.includes('申请') || text.includes('详情')) {
            actions.push({
                text: text.trim().substring(0, 20),
                tag: el.tagName,
                className: el.className
            });
        }
    });
    
    result.actions = actions.slice(0, 10);
    console.log('操作按钮:', actions.length);
    
    // 5. 导出完整数据
    console.log('\n【5. 完整数据】');
    console.log(JSON.stringify(result, null, 2));
    
    // 复制到剪贴板
    navigator.clipboard.writeText(JSON.stringify(result, null, 2)).then(function() {
        console.log('\n✅ 已复制到剪贴板！');
    }).catch(function() {
        console.log('\n⚠️ 请手动复制上面的 JSON');
    });
    
    return result;
})()
