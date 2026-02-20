return {
  {
    "zk-nvim",
    name = "zk.nvim",
    dir = "/home/user/zk-plugin",
    dependencies = {
      "nvim-lua/plenary.nvim",
      "nvim-telescope/telescope.nvim",
    },
    config = function()
      require("zk").setup({ bin = "/home/user/.local/bin/zk" })
    end,
  },
}
