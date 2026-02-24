return {
  {
    dir = "/home/user/zk-plugin",
    name = "zk.nvim",
    lazy = false,
    dependencies = {
      "nvim-lua/plenary.nvim",
      "nvim-telescope/telescope.nvim",
    },
    config = function()
      require("zk").setup({ bin = "/home/user/.local/bin/zk" })
    end,
  },
}
