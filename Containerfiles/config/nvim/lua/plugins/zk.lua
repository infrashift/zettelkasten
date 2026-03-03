return {
  {
    dir = "/home/user/zk-plugin",
    name = "zk.nvim",
    lazy = false,
    dependencies = {
      "nvim-lua/plenary.nvim",
      "folke/snacks.nvim",
    },
    config = function()
      require("zk").setup({ bin = "/home/user/.local/bin/zk" })
    end,
  },
}
