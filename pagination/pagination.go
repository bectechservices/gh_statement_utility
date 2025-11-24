package pagination

import (
	"net/url"
	"strconv"
)

type Pagination struct {
	Limit      int         `json:"limit,omitempty;query:limit"`
	Page       int         `json:"page,omitempty;query:page"`
	Sort       string      `json:"sort,omitempty;query:sort"`
	TotalRows  int64       `json:"total_rows"`
	TotalPages int         `json:"total_pages"`
	Rows       interface{} `json:"rows"`
}

func (p *Pagination) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}
func (p *Pagination) GetLimit() int {
	if p.Limit == 0 {
		p.Limit = 10
	}
	return p.Limit
}
func (p *Pagination) GetPage() int {
	if p.Page == 0 {
		p.Page = 1
	}
	return p.Page
}
func (p *Pagination) GetSort() string {
	if p.Sort == "" {
		p.Sort = "Id desc"
	}
	return p.Sort
}

// Options is a map used to configure tags
type Options map[string]interface{}

//Tag generates the pagination html Tag
func (p Pagination) Tag(opts Options) (string, error) {
	// return a disabled previous and next if there is only 1 page
	if p.TotalPages <= 1 {
		return `<div class="contentFooter-end d-flex align-items-center justify-content-start flex-0-0--auto">
					<div class="ButtonGroup box-root">
						<div class="d-flex align-items-center justify-content-start flex-nowrap"
							 style="margin-left:-8px;margin-top:-8px;">
							<div class="margin-top--8 margin-left--8">
								<div class="box-root">
									<button class="btn btn-sm btn-light pressureButton" disabled>
										<span class="font-weight-medium text-color--dark">Previous</span>
									</button>
								</div>
							</div>
							<div class="margin-top--8 margin-left--8">
								<div class="box-root">
									<button class="btn btn-sm btn-light pressureButton" disabled>
										<span class="font-weight-medium text-color--dark">Next</span>
									</button>
								</div>
							</div>
						</div>
					</div>
				</div>`, nil
	}

	path := extractBaseOptions(opts)

	nextLink, _ := urlFor(path, p.Page+1)

	if p.Page == 1 {
		return `<div class="contentFooter-end d-flex align-items-center justify-content-start flex-0-0--auto">
					<div class="ButtonGroup box-root">
						<div class="d-flex align-items-center justify-content-start flex-nowrap"
							 style="margin-left:-8px;margin-top:-8px;">
							<div class="margin-top--8 margin-left--8">
								<div class="box-root">
									<button class="btn btn-sm btn-light pressureButton" disabled>
										<span class="font-weight-medium text-color--dark">Previous</span>
									</button>
								</div>
							</div>
							<div class="margin-top--8 margin-left--8">
								<div class="box-root">
									<a class="btn btn-sm btn-light pressureButton" href="` + nextLink + `">
										<span class="font-weight-medium text-color--dark">Next</span>
									</a>
								</div>
							</div>
						</div>
					</div>
				</div>`, nil
	}
	previousLink, _ := urlFor(path, p.Page-1)

	if p.Page == p.TotalPages {
		return `<div class="contentFooter-end d-flex align-items-center justify-content-start flex-0-0--auto">
					<div class="ButtonGroup box-root">
						<div class="d-flex align-items-center justify-content-start flex-nowrap"
							 style="margin-left:-8px;margin-top:-8px;">
							<div class="margin-top--8 margin-left--8">
								<div class="box-root">
									<a class="btn btn-sm btn-light pressureButton" href="` + previousLink + `" >
										<span class="font-weight-medium text-color--dark">Previous</span>
									</a>
								</div>
							</div>
							<div class="margin-top--8 margin-left--8">
								<div class="box-root">
									<button class="btn btn-sm btn-light pressureButton" disabled>
										<span class="font-weight-medium text-color--dark">Next</span>
									</button>
								</div>
							</div>
						</div>
					</div>
				</div>`, nil
	}
	//load all with pre and next enabled

	return `<div class="contentFooter-end d-flex align-items-center justify-content-start flex-0-0--auto">
					<div class="ButtonGroup box-root">
						<div class="d-flex align-items-center justify-content-start flex-nowrap"
							 style="margin-left:-8px;margin-top:-8px;">
							<div class="margin-top--8 margin-left--8">
								<div class="box-root">
									<a class="btn btn-sm btn-light pressureButton" href="` + previousLink + `" >
										<span class="font-weight-medium text-color--dark">Previous</span>
									</a>
								</div>
							</div>
							<div class="margin-top--8 margin-left--8">
								<div class="box-root">
									<a class="btn btn-sm btn-light pressureButton" href="` + nextLink + `" >
										<span class="font-weight-medium text-color--dark">Next</span>
									</a>
								</div>
							</div>
						</div>
					</div>
				</div>`, nil
}

func extractBaseOptions(opts Options) string {
	var path string
	if p, ok := opts["path"]; ok {
		path = p.(string)
		delete(opts, "path")
	}

	return path
}

func urlFor(path string, page int) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("page", strconv.Itoa(page))
	u.RawQuery = q.Encode()

	return u.String(), err
}
