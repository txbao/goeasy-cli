package generator

func metaForTest(id, domain, resource string) ModuleMeta {
	m := ModuleMeta{
		ModuleID: id, ModuleSnake: id,
		Domain: domain, Resource: resource,
		Group: domain, Grouped: true,
	}
	normalizeMeta(&m)
	return m
}
