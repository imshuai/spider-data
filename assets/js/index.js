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
            currentPage: 1,
            data: [],
            total: 0,
            project: "",
            loading: false
        }
    },
    methods: {
        getData: function () {
            if (this.project == "") {
                return
            }
            if (this.currentPage < 1) {
                this.currentPage = 1
            }
            this.loading = true;
            console.log(this.currentPage);
            axios.get("./api/" + this.project + "/" + this
                    .currentPage, {
                        params: {
                            num: this.limit
                        }
                    })
                .then((response) => {
                    this.data = response.data.data;
                    this.total = parseInt(response.data.total);
                    this.loading = false;
                })
                .catch((error) => {
                    this.loading = false;
                    this.$message.error(error);
                })
        },
        sizeChange: function (val) {
            this.limit = val;
            this.currentPage = 1;
            this.data = [];
            this.getData();
        },
        currentChange: function (val) {
            this.currentPage = val;
            localStorage.setItem("cPage", val)
            this.data = [];
            this.getData();
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
                axios.delete('./api/' + this.project + "/" + md5(row.name))
                    .then(res => {
                        this.$notify({
                            'type': 'success',
                            'title':'提示',
                            'message': '删除 ' + row.name + ' ' + res.data.status,
                            'duration':1500
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
                    "type":"info"
                });
            });

        },
        projectChange: function () {
            this.getData();
        }
    },
    mounted() {
        var page = localStorage.getItem("cPage");
        if (page == null) {
            page = "1"
        }
        this.currentPage = parseInt(page);
        this.getData();
    }
});