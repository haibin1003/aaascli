// ============================================
// 脚本 4：导出订购/授权流程数据
// 在点击"订购"/"授权"按钮后出现弹窗/页面时执行
// ============================================

(function() {
    console.log('=== 订购/授权流程数据导出 ===\n');
    
    var result = {
        exportTime: new Date().toISOString(),
        pageUrl: window.location.href,
        pageTitle: document.title,
        step: 'order-dialog' // 标记这是订购流程
    };
    
    // 1. 弹窗/对话框标题
    console.log('【1. 弹窗/页面标题】');
    var dialogTitle = document.querySelector('.el-dialog__title, .modal-title, .dialog-title, h2, h3');
    result.dialogTitle = dialogTitle ? dialogTitle.innerText.trim() : '';
    console.log('标题:', result.dialogTitle);
    
    // 2. 表单字段详情
    console.log('\n【2. 表单字段（重点）】');
    result.formFields = [];
    
    document.querySelectorAll('input, select, textarea, .el-input, .el-select').forEach(function(el) {
        var label = '';
        var placeholder = el.getAttribute('placeholder') || '';
        
        // 找 label
        var id = el.getAttribute('id');
        var name = el.getAttribute('name');
        
        if (id) {
            var labelEl = document.querySelector('label[for="' + id + '"]');
            if (labelEl) label = labelEl.innerText;
        }
        
        // 如果找不到，找相邻元素
        if (!label) {
            var parent = el.parentElement;
            if (parent) {
                var labelInParent = parent.querySelector('label');
                if (labelInParent) label = labelInParent.innerText;
            }
        }
        
        // 查找必填标记
        var isRequired = el.getAttribute('required') !== null ||
                         el.classList.contains('is-required') ||
                         (label && label.includes('*'));
        
        result.formFields.push({
            tag: el.tagName,
            type: el.type || '',
            name: name || '',
            label: label,
            placeholder: placeholder,
            isRequired: isRequired,
            value: el.value || ''
        });
    });
    
    console.log('找到', result.formFields.length, '个表单字段');
    result.formFields.forEach(function(f) {
        console.log('  -', f.label || f.placeholder || f.name, f.isRequired ? '(必填)' : '');
    });
    
    // 3. 选项/下拉菜单选项
    console.log('\n【3. 下拉选项】');
    result.selectOptions = [];
    document.querySelectorAll('select, .el-select').forEach(function(el) {
        var label = '';
        var parent = el.parentElement;
        if (parent) {
            var labelEl = parent.querySelector('label');
            if (labelEl) label = labelEl.innerText;
        }
        
        var options = [];
        el.querySelectorAll('option, .el-select-dropdown__item').forEach(function(opt) {
            options.push(opt.innerText.trim());
        });
        
        if (options.length > 0) {
            result.selectOptions.push({
                label: label,
                options: options
            });
        }
    });
    console.log('找到', result.selectOptions.length, '个下拉菜单');
    
    // 4. 步骤/流程指示器
    console.log('\n【4. 订购流程步骤】');
    result.steps = [];
    document.querySelectorAll('.step, .el-step, .process-step, [class*="step"]').forEach(function(el) {
        var text = el.innerText || el.textContent || '';
        if (text.trim() && text.trim().length < 30) {
            result.steps.push({
                text: text.trim(),
                className: el.className,
                isActive: el.classList.contains('is-active') || el.classList.contains('active'),
                isComplete: el.classList.contains('is-complete') || el.classList.contains('complete')
            });
        }
    });
    console.log('流程步骤:', result.steps.map(function(s) { return s.text; }).join(' → ') || '未找到');
    
    // 5. 按钮（提交/取消）
    console.log('\n【5. 操作按钮】');
    result.buttons = [];
    document.querySelectorAll('button, .btn, [class*="button"]').forEach(function(el) {
        var text = el.innerText || el.textContent || '';
        if (text.includes('提交') || text.includes('确认') || text.includes('取消') || 
            text.includes('下一步') || text.includes('上一步') || text.includes('保存')) {
            result.buttons.push({
                text: text.trim(),
                type: el.type || 'button',
                className: el.className
            });
        }
    });
    console.log('找到按钮:', result.buttons.map(function(b) { return b.text; }).join(', '));
    
    // 6. 提示信息
    console.log('\n【6. 提示/说明信息】');
    result.tips = [];
    document.querySelectorAll('.tip, .tips, .notice, .alert, .el-alert, [class*="tip"]').forEach(function(el) {
        var text = el.innerText || el.textContent || '';
        if (text.trim() && text.trim().length < 200) {
            result.tips.push(text.trim());
        }
    });
    console.log('找到', result.tips.length, '条提示');
    
    // 7. Vue 表单数据
    console.log('\n【7. Vue 表单数据】');
    result.vueForm = null;
    var allElements = document.querySelectorAll('*');
    
    for (var i = 0; i < allElements.length; i++) {
        var vue = allElements[i].__vue__;
        if (vue && vue.$data) {
            var keys = Object.keys(vue.$data);
            var formKeys = keys.filter(function(k) {
                var val = vue.$data[k];
                return (k.toLowerCase().includes('form') || 
                        k.toLowerCase().includes('model') ||
                        k.toLowerCase().includes('data')) &&
                       typeof val === 'object' && 
                       val !== null &&
                       Object.keys(val).length > 0;
            });
            
            if (formKeys.length > 0) {
                result.vueForm = {
                    keys: keys,
                    formKeys: formKeys,
                    formData: {}
                };
                
                formKeys.forEach(function(k) {
                    var val = vue.$data[k];
                    result.vueForm.formData[k] = {
                        keys: Object.keys(val),
                        values: JSON.stringify(val).substring(0, 800)
                    };
                });
                
                break;
            }
        }
    }
    
    if (result.vueForm) {
        console.log('找到 Vue 表单数据');
    } else {
        console.log('未找到 Vue 表单数据');
    }
    
    // 8. 页面完整文本
    console.log('\n【8. 页面文本预览】');
    result.pageText = document.body.innerText.substring(0, 1500);
    
    // 9. 导出结果
    console.log('\n=== 导出完成 ===');
    console.log(JSON.stringify(result, null, 2));
    
    navigator.clipboard.writeText(JSON.stringify(result, null, 2)).then(function() {
        console.log('\n✅ 已复制到剪贴板！请粘贴给我');
    }).catch(function() {
        console.log('\n⚠️ 自动复制失败，请手动复制上面的 JSON');
    });
    
    return result;
})();
