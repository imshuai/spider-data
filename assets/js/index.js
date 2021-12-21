var app = new Vue({
    el: '#app',
    data() {
        return {
            cols: [{
                id: "name",
                label: "文件名",
                width: "100px"
            }, {
                id: "size",
                label: "文件大小",
                width: "auto"
            }, {
                id: "last-modified",
                label: "最后修改时间",
                width: "auto"
            }],
            limit: 20,
            data: [],
            total: 0,
            currentPage: 1,
            firstkey: "",
            lastkey: "",
            project: "",
            loading: false
        }
    },
    methods: {
        getData: function (act, key) {
            if (this.project == "") {
                return
            }
            this.loading = true;
            axios.get("/api/" + this.project, {
                    params: {
                        act: act,
                        num: this.limit,
                        key: key
                    }
                })
                .then((response) => {
                    this.data = response.data.data;
                    this.total = parseInt(response.data.total);
                    this.firstkey = this.project + ":data:" + md5(this.data[0].name);
                    this.lastkey = this.project + ":data:" + md5(this.data[this.data.length - 1].name);
                    this.loading = false;
                })
                .catch((error) => {
                    this.loading = false;
                    this.$message.error(error);
                })
        },
        sizeChange: function (val) {
            this.limit = val;
            localStorage.setItem("limit", this.limit);
            this.data = [];
            this.getData("mount", this.firstkey);
        },
        currentChange: function (val) {
            if (val < parseInt(localStorage.getItem("currentPage"))) {
                this.prevClick(val);
            } else {
                this.nextClick(val);
            }
        },
        prevClick: function (val) {
            this.getData("prev", this.firstkey);
            this.currentPage = val;
            localStorage.setItem("firstkey", this.firstkey);
            localStorage.setItem("lastkey", this.lastkey);
            localStorage.setItem("currentPage", val);
        },
        nextClick: function (val) {
            this.getData("next", this.lastkey);
            this.currentPage = val;
            localStorage.setItem("firstkey", this.firstkey);
            localStorage.setItem("lastkey", this.lastkey);
            localStorage.setItem("currentPage", val);
        },
        handleClick: function (row) {
            window.open('javascript:window.name;', '<script>location.replace("' + row
                .url + '")<\/script>');
        },
        handleDelete: function (idx, row) {
            this.$confirm();
            this.$confirm('确定删除 ' + row.name + ' ?', '警告', {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
                type: 'warning',
            }).then(action => {
                axios.delete('/api/' + this.project + "/" + md5(row.name))
                    .then(res => {
                        this.$notify({
                            'type': 'success',
                            'title': '提示',
                            'message': '删除 ' + row.name + ' ' + res.data.status,
                            'duration': 1500
                        });
                        if (res.data.status == "成功") {
                            this.data.splice(idx, 1);
                            this.total--;
                        }
                    })
                    .catch(err => {
                        this.$message.error(err);
                    })
            }).catch(() => {
                this.$notify({
                    "title": "提示",
                    "message": "操作取消",
                    "type": "info"
                });
            });

        },
        projectChange: function () {
            this.firstkey = "";
            this.lastkey = "";
            localStorage.setItem("project", this.project);
            localStorage.setItem("firstkey", this.firstkey);
            localStorage.setItem("lastkey", this.lastkey);
            this.getData("mount", this.firstkey);
        }
    },
    mounted() {
        var project = localStorage.getItem("project");
        if (project == null) {
            project = ""
        }
        this.project = project;
        var limit = localStorage.getItem("limit");
        if (limit == null) {
            limit = 20
        }
        this.limit = parseInt(limit);
        var page = localStorage.getItem("currentPage");
        if (page == null) {
            page = 1
        }
        this.currentPage = parseInt(page);
        var firstkey = localStorage.getItem("firstkey");
        if (firstkey == null) {
            firstkey = ""
        }
        this.firstkey = firstkey;
        var lastkey = localStorage.getItem("lastkey");
        if (lastkey == null) {
            lastkey = ""
        }
        this.lastkey = lastkey;
        this.getData("mount", this.firstkey);
    }
});