// admin.js
window.onload = function() {
    const taskTable = document.getElementById('task-table').getElementsByTagName('tbody')[0];
    const token = localStorage.getItem('token'); // 获取存储的管理员token

    // 检查token是否存在
    if (!token) {
        alert("请先登录！");
        window.location.href = 'index.html';  // 跳转到登录页面
        return;
    }

    // 获取待审批任务
    fetch('/admin/pending_tasks', {
        method: 'GET',
        headers: {
            'Authorization': `Bearer ${token}` // 将 token 作为请求头传递
        }
    })
        .then(response => response.json())
        .then(data => {
            if (data.tasks) {
                data.tasks.forEach(task => {
                    const row = taskTable.insertRow();
                    row.innerHTML = `
                    <td>${task.id}</td>
                    <td>${task.target_url}</td>
                    <td>${task.start_time}</td>
                    <td>${task.end_time}</td>
                    <td><button class="approve-btn" onclick="approveTask(${task.id})">审批</button></td>
                `;
                });
            } else {
                alert("没有待审批任务");
            }
        })
        .catch(error => {
            console.error('Error:', error);
            alert('获取待审批任务失败');
        });
};

// 审批任务
function approveTask(taskId) {
    const token = localStorage.getItem('token');

    fetch('/admin/approve_loadtest', {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: `id=${taskId}`
    })
        .then(response => response.json())
        .then(data => {
            if (data.message === '任务审批成功') {
                alert('任务审批成功');
                location.reload(); // 刷新页面，重新加载待审批任务
            } else {
                alert('审批失败');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            alert('审批失败');
        });
}
