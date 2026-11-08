# frozen_string_literal: true

require "addressable/uri"
require "nokogiri"
require "pathname"

# The relative-links plugin handles published pages. Supporting files stay on GitHub.
Jekyll::Hooks.register :pages, :post_render do |page|
  next unless page.ext == ".md"

  config = page.site.config
  project = Pathname.new(config.fetch("ticket_project"))
  plan = project.join(config.fetch("ticket_plan"))
  source = plan.join(page.relative_path.delete_prefix("roadmap/"))
  published_prefix = "#{config.fetch('baseurl', '')}/roadmap/"
  document = Nokogiri::HTML(page.output)
  document.css("a[href], img[src]").each do |node|
    attribute = node.name == "a" ? "href" : "src"
    link = Addressable::URI.parse(node[attribute])
    next if link.scheme || link.host || link.path.to_s.empty? || link.path.start_with?(published_prefix)

    path = Addressable::URI.unencode(link.path)
    target = (path.start_with?("/") ? project.join(path.delete_prefix("/")) : source.dirname.join(path)).cleanpath
    relative = target.relative_path_from(project).to_s
    raise "Supporting link escapes repository: #{node[attribute]}" if relative == ".." || relative.start_with?("../")

    location = node.name == "img" ? "https://raw.githubusercontent.com" : "https://github.com"
    revision = Addressable::URI.encode_component(config.fetch("ticket_revision"), Addressable::URI::CharacterClasses::UNRESERVED)
    prefix = "#{location}/#{config.fetch('ticket_repository')}/#{node.name == 'a' ? 'blob/' : ''}#{revision}/"
    encoded_path = relative.split("/").map do |component|
      Addressable::URI.encode_component(component, Addressable::URI::CharacterClasses::UNRESERVED)
    end.join("/")
    target_url = Addressable::URI.parse(prefix + encoded_path)
    target_url.query, target_url.fragment = link.query, link.fragment
    node[attribute] = target_url.to_s
  end
  page.output = document.to_html
end
