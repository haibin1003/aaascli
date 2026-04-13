// ============================================
// 脚本 2：导出能力/服务详情页数据
// 在进入任意能力详情页后执行
// ============================================

(function() {
    console.log('=== 能力详情页数据导出 ===\n');
    
    var result = {
        exportTime: new Date().toISOString(),
        pageUrl: window.location.href,
        pageTitle: document.title
    };
    
    // 1. 页面基本信息
    console.log('【1. 页面信息】');
    console.log('URL:', result.pageUrl);
    console.log('标题:', result.pageTitle);
    
    // 2. 能力基本信息（从页面文本提取）
    console.log('\n【2. 能力信息】');
    result.abilityInfo = {
        name: '',
        code: '',
        provider: '',
        description: '',
        status: ''
    };
    
    // 尝试提取能力名称
    var titleEl = document.querySelector('h1, h2, .title, [class*="title"]');
    if (titleEl) {
        result.abilityInfo.name = titleEl.innerText.trim();
    }
    
    // 尝试提取提供方
    var bodyText = document.body.innerText;
    var providerMatch = bodyText.match(/提供方[：:]\s*(.+?)(?:\n|$)/);
    if (providerMatch) {
        result.abilityInfo.provider = providerMatch[1].trim();
    }
    
    // 尝试提取能力编号
    var codeMatch = bodyText.match(/能力编号[：:]\s*(.+?)(?:\n|$)/) || 
                    bodyText.match(/编号[：:]\s*(.+?)(?:\n|$)/);
    if (codeMatch) {
        result.abilityInfo.code = codeMatch[1].trim();
    }
    
    console.log('能力名称:', result.abilityInfo.name);
    console.log('提供方:', result.abilityInfo.provider);
    console.log('能力编号:', result.abilityInfo.code);
    
    // 3. 标签页/选项卡
    console.log('\n【3. 详情标签页】');
    result.tabs = [];
    document.querySelectorAll('.tab, .el-tabs__item, .nav-tab, [role="tab"]').forEach(function(el) {
        var text = el.innerText || el.textContent || '';
        if (text.trim() && text.trim().length < 20) {
            result.tabs.push({
                name: text.trim(),
                className: el.className,
                isActive: el.classList.contains('is-active') || el.classList.contains('active')
            });
        }
    });
    console.log('找到', result.tabs.length, '个标签:', result.tabs.map(function(t) { return t.name; }));
    
    // 4. 关键按钮（订购、授权、申请等）
    console.log('\n【4. 关键操作按钮】');
    result.actionButtons = [];
    document.querySelectorAll('button, a, .btn').forEach(function(el) {
        var text = el.innerText || el.textContent || '';
        var lowerText = text.toLowerCase();
        
        if (text.includes('订购') || text.includes('授权') || text.includes('申请') || 
            text.includes('使用') || text.includes('开通') || text.includes('订阅') ||
            text.includes('立即') || text.includes('点击') || text.includes('我要')) {
            
            result.actionButtons.push({
                text: text.trim(),
                tag: el.tagName,
                className: el.className,
                id: el.id,
                href: el.getAttribute('href') || '',
                onclick: el.getAttribute('onclick') || ''
            });
        }
    });
    
    console.log('找到', result.actionButtons.length, '个操作按钮');
    result.actionButtons.forEach(function(b) {
        console.log('  -', b.text, '(' + b.tag + ')');
    });
    
    // 5. 表单字段（如果有）
    console.log('\n【5. 表单字段】');
    result.formFields = [];
    document.querySelectorAll('input, select, textarea, .el-input').forEach(function(el) {
        var placeholder = el.getAttribute('placeholder') || '';
        var label = '';
        
        // 找对应的 label
        var id = el.getAttribute('id');
        if (id) {
            var labelEl = document.querySelector('label[for="' + id + '"]');
            if (labelEl) label = labelEl.innerText;
        }
        
        if (placeholder || label) {
            result.formFields.push({
                type: el.type || el.tagName,
                placeholder: placeholder,
                label: label,
                name: el.getAttribute('name') || ''
            });
        }
    });
    console.log('找到', result.formFields.length, '个表单字段');
    
    // 6. 价格信息
    console.log('\n【6. 价格信息】');
    result.priceInfo = '';
    var bodyText = document.body.innerText;
    var priceMatch = bodyText.match(/价格[：:]\s*(.+?)(?:\n|$)/) ||
                     bodyText.match(/费用[：:]\s*(.+?)(?:\n|$)/) ||
                     bodyText.match(/¥\d+/) ||
                     bodyText.match(/\d+元/);
    if (priceMatch) {
        result.priceInfo = priceMatch[0];
        console.log('价格:', result.priceInfo);
    } else {
        console.log('未找到价格信息');
    }
    
    // 7. Vue 详情数据
    console.log('\n【7. Vue 详情数据】');
    result.vueDetail = null;
    var allElements = document.querySelectorAll('*');
    
    for (var i = 0; i < allElements.length; i++) {
        var vue = allElements[i].__vue__;
        if (vue && vue.$data) {
            var keys = Object.keys(vue.$data);
            // 查找包含 detail/info/data 的对象
            var detailKeys = keys.filter(function(k) {
                var val = vue.$data[k];
                return (k.toLowerCase().includes('detail') || 
                        k.toLowerCase().includes('info') ||
                        k.toLowerCase().includes('data')) &&
                       typeof val === 'object' && 
                       val !== null && 
                       !Array.isArray(val) &&
                       Object.keys(val).length > 0;
            });
            
            if (detailKeys.length > 0) {
                result.vueDetail = {
                    keys: keys,
                    detailKeys: detailKeys,
                    data: {}
                };
                
                detailKeys.forEach(function(k) {
                    var val = vue.$data[k];
                    result.vueDetail.data[k] = {
                        keys: Object.keys(val),
                        sample: JSON.stringify(val).substring(0, 500)
                    };
                });
                
                break;
            }
        }
    }
    
    if (result.vueDetail) {
        console.log('找到 Vue 详情数据');
    } else {
        console.log('未找到 Vue 详情数据');
    }
    
    // 8. 页面文本预览（前1000字符）
    console.log('\n【8. 页面文本预览】');
    result.pageTextPreview = document.body.innerText.substring(0, 1000);
    
    // 9. 导出完整结果
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
