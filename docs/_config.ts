import lume from "lume/mod.ts";
import basePath from "lume/plugins/base_path.ts";

const site = lume({
  src: ".",
  dest: "_site",
  location: new URL("https://exists.lol/"),
});

site.use(basePath());

site.copy("assets");

export default site;
